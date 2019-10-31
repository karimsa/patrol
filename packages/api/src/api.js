import * as path from 'path'

import express from 'express'
import morgan from 'morgan'
import cors from 'cors'

import { model } from './db'

const NO_RESPONSE = Symbol('NO_RESPONSE')

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
				.findOne({
					service: serviceCheck.service,
					check: serviceCheck.check.name,
				})
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

export function createApp(config) {
	const app = express()

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
					`https://img.shields.io/badge/patrol-down-red?style=${style}`,
				)
			} else {
				res.redirect(
					`https://img.shields.io/badge/patrol-up-brightgreen?style=${style}`,
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
			if (typeof $limit !== 'number') {
				throw new APIError(
					`'$limit' must be provided in query, and must be a valid integer`,
					400,
				)
			}

			return model('Checks').find(
				{
					service,
					check,
				},
				{
					limit: $limit,
				},
			)
		}),
	)

	app.get(
		'/api/checks',
		route(async () =>
			(await getStatusChecks(checkList)).reduce((groups, check) => {
				groups[check.service] = groups[check.service] || []
				groups[check.service].push(check)
				return groups
			}, {}),
		),
	)

	return app
}
