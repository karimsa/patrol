import { promises as fs } from 'fs'
import * as zlib from 'zlib'
import * as path from 'path'
import * as http from 'http'

import cheerio from 'cheerio'
import express from 'express'
import morgan from 'morgan'
import cors from 'cors'
import initSocketIO from 'socket.io'
import { logger } from '@karimsa/boa'
import compression from 'compression'

import { model } from './db'

const NO_RESPONSE = Symbol('NO_RESPONSE')

// For SSR
async function render(serviceChecksMap) {
	// Render <App /> as static markup
	process.env.IS_SERVER = 'true'
	const DEFAULT_STATE = {
		checks: {
			status: 'success',
			result: serviceChecksMap,
		},
		checkHistory: {},
	}
	global.window = {
		DEFAULT_STATE,
	}
	const app = require('../web/dist/server')
	global.window = null
	const html = await fs.readFile(
		path.resolve(__dirname, '..', 'web', 'dist', 'index.html'),
		'utf8',
	)
	const $ = cheerio.load(html)
	$('#app').html(app.HTML)

	// Inline css
	const stylesheets = $('link[rel="stylesheet"]')
	for (let i = 0; i < stylesheets.length; ++i) {
		const cssFile =
			path.resolve(__dirname, '..', 'web', 'dist') +
			$(stylesheets[i]).attr('href')
		const css = await fs.readFile(cssFile)
		$(stylesheets[i]).replaceWith(`<style>${css}</style>`)
	}

	// Render redux initial state
	const files = await fs.readdir(path.resolve(__dirname, '..', 'web', 'dist'))
	const ssrFile = files.find(f => f.startsWith('ssr') && f.endsWith('.js'))
	if (!ssrFile) {
		throw new Error(`Failed to find ssr.js in ../web/dist`)
	}
	await fs.writeFile(
		path.resolve(__dirname, '..', 'web', 'dist', ssrFile),
		`window.DEFAULT_STATE = ${JSON.stringify(DEFAULT_STATE)}`,
	)

	const staticHTML = $.html()
	const brotliHTML = await new Promise((resolve, reject) => {
		zlib.brotliCompress(staticHTML, (err, result) => {
			if (err) reject(err)
			else resolve(result)
		})
	})

	return {
		raw: staticHTML,
		brotliHTML,
	}
}

function route(fn) {
	return function(req, res) {
		fn(req, res)
			.then(body => {
				if (body !== NO_RESPONSE) {
					res.json(body)
				}
			})
			.catch(error => {
				res.status(500)
				res.json({
					error: String(error.stack || error),
				})
			})
	}
}

class APIError extends Error {
	constructor(message, status) {
		super(message)
		this.status = status
	}
}

function getStatusChecks(checkList) {
	return Promise.all(
		checkList.map(serviceCheck => {
			return model('Checks')
				.findOne(
					{
						service: serviceCheck.service,
						check: serviceCheck.check.name,
					},
					{
						sort: {
							createdAt: -1,
						},
					},
				)
				.then(check => {
					return (
						check || {
							_id: [
								'todo',
								Date.now(),
								serviceCheck.service,
								serviceCheck.check.name,
							].join('-'),
							service: serviceCheck.service,
							check: serviceCheck.check.name,
							serviceStatus: 'inprogress',
							createdAt: Date.now(),
							output: '',
						}
					)
				})
		}),
	)
}

async function getServiceMap(checkList) {
	return (await getStatusChecks(checkList)).reduce((groups, check) => {
		groups[check.service] = groups[check.service] || []
		groups[check.service].push(check)
		return groups
	}, {})
}

export const io = {
	listeners: [],
	emit(event, data) {
		for (const fn of this.listeners) {
			fn(event, data)
		}
	},
}

export function createApp(config) {
	const app = express()
	const server = http.createServer(app)
	const socketServer = initSocketIO(server)

	let staticHTML

	socketServer.on('connection', sock => {
		logger.info(`Socket connected`)
		sock.on('close', () => {
			logger.info(`Socket disconnected`)
		})
	})

	io.listeners.push(function(event, data) {
		logger.info(`Broadcasting socket event %O`, event)
		socketServer.emit(event, data)
	})

	app.use(compression())
	app.set('etag', false)
	app.get('/', async (req, res, next) => {
		if ((req.path === '/' || req.path === '/index.html') && staticHTML) {
			if (req.acceptsEncodings('br')) {
				res.set('Content-Encoding', 'br')
				res.end(staticHTML.brotliHTML)
			} else {
				res.end(staticHTML.raw)
			}
			return
		}

		next()
	})
	app.use(express.static(path.resolve(__dirname, '..', 'web', 'dist')))
	app.use(morgan('dev'))

	if (process.env.NODE_ENV !== 'production') {
		app.use(
			cors({
				origin: 'http://localhost:1234',
				credentials: true,
			}),
		)
	}

	const checkList = []

	for (const name in config.services) {
		if (config.services.hasOwnProperty(name)) {
			for (const check of config.services[name].checks) {
				checkList.push({
					service: name,
					check,
				})
			}
		}
	}

	getServiceMap(checkList)
		.then(serviceMap => render(serviceMap))
		.then(html => {
			logger.info(`Rendered static markup`)
			staticHTML = html
		})
		.catch(error => {
			logger.error(`Failed to render static HTML`, error)
		})

	app.get(
		'/badge',
		route(async (req, res) => {
			let numSystemsUnhealthy = 0
			for (const { serviceStatus } of await getStatusChecks(checkList)) {
				if (serviceStatus === 'unhealthy') {
					numSystemsUnhealthy++
				}
			}

			const { style = 'for-the-badge' } = req.query

			if (numSystemsUnhealthy === 0) {
				res.redirect(
					`https://img.shields.io/badge/patrol-up-brightgreen?style=${style}`,
				)
			} else {
				res.redirect(
					`https://img.shields.io/badge/patrol-down-red?style=${style}`,
				)
			}

			return NO_RESPONSE
		}),
	)

	app.get(
		'/api/config',
		route(async () => {
			return {
				title: config.web.title,
			}
		}),
	)

	app.get(
		'/api/checks/history',
		route(async req => {
			const { service, check } = req.query
			if (typeof service !== 'string') {
				throw new APIError(
					`'service' must be provided in query, and must be a string`,
					400,
				)
			}
			if (typeof check !== 'string') {
				throw new APIError(
					`'check' must be provided in query, and must be a string`,
					400,
				)
			}

			const $limit = parseInt(req.query.$limit, 10)
			if (typeof $limit !== 'number' || isNaN($limit)) {
				throw new APIError(
					`'$limit' must be provided in query, and must be a valid integer`,
					400,
				)
			}

			const entries = await model('Checks').find(
				{
					service,
					check,
				},
				{
					limit: $limit,
					sort: {
						createdAt: -1,
					},
				},
			)
			return entries.reverse()
		}),
	)

	app.get('/api/checks', route(async () => getServiceMap(checkList)))

	return {
		app,
		server,
	}
}
