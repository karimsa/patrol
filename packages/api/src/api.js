import * as path from 'path'

import express from 'express'
import morgan from 'morgan'
import cors from 'cors'

import { model } from './db'

function route(fn) {
	return function(req, res) {
		fn(req, res)
			.then(body => res.json(body))
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
			for (const check of config.services[name]) {
				checkList.push({
					service: name,
					check,
				})
			}
		}
	}

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
		route(async () => {
			const checks = await Promise.all(
				checkList.map(serviceCheck => {
					return model('Checks').findOne({
						service: serviceCheck.service,
						check: serviceCheck.check.name,
					}).then(check => {
						return check || {
							service: serviceCheck.service,
							check: serviceCheck.check.name,
							serviceStatus: 'inprogress',
						}
					})
				}),
			)

			const groups = {}
			for (const check of checks) {
				groups[check.service] = groups[check.service] || []
				groups[check.service].push(check)
			}
			return groups
		}),
	)

	return app
}
