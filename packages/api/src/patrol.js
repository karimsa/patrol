import * as fs from 'fs'
import * as path from 'path'
import * as util from 'util'
import * as os from 'os'
import http from 'http'

import ms from 'ms'
import { logger } from '@karimsa/boa'
import yargs from 'yargs'
import yaml from 'js-yaml'

import { initDB } from './db'
import { startWithConfig } from './checker'
import * as queue from './queue'
import { normalizeNotifications } from './notifiers'
import { createApp } from './api'

const readFile = util.promisify(fs.readFile)
const defaultWebConfig = {
	port: 1234,
	apiPort: 8080,
	title: 'System Status',
}

async function main() {
	const argv = yargs
		.option('config', {
			alias: 'c',
			type: 'string',
			describe: 'Absolute path to your patrol config file',
			demand: true,
		})
		.option('concurrency', {
			alias: 'n',
			type: 'number',
			describe: 'Number of operations to run concurrently',
			default: os.cpus().length,
		}).argv

	if (argv._.length !== 0) {
		console.error(`Unrecognized commands: [${argv._}]`)
		yargs.showHelp()
	}

	if (argv.config[0] !== '/') {
		console.error(`Config file must be an absolute path`)
		yargs.showHelp()
		return
	}

	// Parse the config file
	const configData = await readFile(argv.config, 'utf8')
	const config = yaml.safeLoad(configData)

	// If the path to the db is missing, put it next to the
	// config file
	if (!config.dbDirectory) {
		config.dbDirectory = path.resolve(path.dirname(argv.config), 'db')
	}

	// Verify the 'services' object
	if (typeof config.services !== 'object' || Array.isArray(config.services)) {
		console.error(
			`'services' property must be provided at the top-level, as a dictionary`,
		)
		yargs.showHelp()
		return
	}

	let hasErrors = false

	// Normalize notifications
	if (config.notifications) {
		if (typeof config.notifications !== 'object') {
			console.error(`Error: 'notifications' should be a dictionary`)
			hasErrors = true
		} else {
			if (config.notifications.on_success) {
				const errors = normalizeNotifications(config.notifications.on_success)
				if (errors) {
					console.error(errors)
					hasErrors = true
				}
			}
		}
	} else {
		config.notifications = {}
	}

	for (const name in config.services) {
		if (config.services.hasOwnProperty(name)) {
			if (Array.isArray(config.services[name])) {
				config.services[name] = {
					checks: config.services[name],
					notifications: undefined,
				}
			}

			if (
				typeof config.services[name] !== 'object' ||
				!Array.isArray(config.services[name].checks)
			) {
				throw new Error(
					`Service should be either an array of checks or have a '.checks' key with an array of checks`,
				)
			}

			if (config.services[name].notifications) {
				if (config.services[name].notifications.on_success) {
					normalizeNotifications(config.services[name].notifications.on_success)
					if (config.notifications.on_success) {
						config.services[name].notifications.on_success.push(
							...config.notifications.on_success,
						)
					}
				}
				if (config.services[name].notifications.on_failure) {
					normalizeNotifications(config.services[name].notifications.on_failure)
					if (config.notifications.on_failure) {
						config.services[name].notifications.on_failure.push(
							...config.notifications.on_failure,
						)
					}
				}
			} else {
				config.services[name].notifications = config.notifications
			}

			for (
				let index = 0;
				index < config.services[name].checks.length;
				++index
			) {
				const check = config.services[name].checks[index]

				if (typeof check.name !== 'string' || !check.name) {
					console.error(
						`Error: 'services.${name}[${index}].name' must be a valid string (got: ${JSON.stringify(
							check.name,
						)})`,
					)
					hasErrors = true
				}
				if (Array.isArray(check.cmd)) {
					check.cmd = check.cmd.join('; ')
				}
				if (typeof check.cmd !== 'string' || !check.cmd) {
					console.error(
						`Error: 'services.${name}[${index}].cmd' must be a valid string (got: ${JSON.stringify(
							check.cmd,
						)})`,
					)
					hasErrors = true
				}

				// Normalize `check.interval`, which can either be a direct
				// number of milliseconds or an ms-recognized string such as `10s`
				if (check.interval === undefined) {
					check.interval = 60 * 1000
				} else if (typeof check.interval === 'string') {
					check.interval = ms(check.interval)
				}
				if (typeof check.interval !== 'number') {
					console.error(
						`Error: 'services.${name}[${index}].interval' is not a valid time interval`,
					)
					hasErrors = true
				}

				// Normalize image, leaving validation up to docker
				// Defaulting to custom image
				if (!check.image) {
					check.image = 'byrnedo/alpine-curl'
				} else if (typeof check.image !== 'string') {
					console.error(
						`Error: 'services.${name}[${index}].image' must be a valid docker image`,
					)
					hasErrors = true
				}
			}
		}
	}

	// Normalize web
	if (!config.web) {
		config.web = defaultWebConfig
	}

	if (!config.web.title) {
		config.web.title = defaultWebConfig.title
	}
	if (typeof config.web.title !== 'string') {
		console.error(`web.title should be a string`)
		hasErrors = true
	}

	if (hasErrors) {
		console.error()
		yargs.showHelp()
		return
	}

	// Initialize collections
	initDB(config.dbDirectory)

	// Create the API server
	const app = createApp(config)
	const server = http.createServer(app)
	await new Promise((resolve, reject) => {
		server.on('error', reject)
		server.listen(config.port || 8080, resolve)
	})
	logger.info(`Started patrol API server on :%O`, server.address().port)

	// Start the first scan
	await queue.Enqueue(() =>
		startWithConfig({
			config,
		}),
	)

	// Run the worker
	await queue.PerformWork({
		concurrency: argv.concurrency,
	})
}

main().catch(error => {
	console.error(error.stack)
	process.exit(1)
})
