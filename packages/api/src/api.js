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

export function createApp(config) {
	const app = express()

	app.use(morgan('dev'))
	app.use(cors({
		origin: 'http://localhost:1234',
		credentials: true,
	}))

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

	app.get('/api/checks', route(async () => {
		const checks = await Promise.all(
			checkList.map(serviceCheck => {
				return model('Checks').findOne({
					service: serviceCheck.service,
					check: serviceCheck.check.name,
				})
			})
		)

		const groups = {}
		for (const check of checks) {
			groups[check.service] = groups[check.service] || []
			groups[check.service].push(check)
		}
		return groups
	}))

	return app
}
