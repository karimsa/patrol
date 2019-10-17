import * as fs from 'fs'
import * as path from 'path'
import * as util from 'util'
import * as os from 'os'
import ms from 'ms'

import yargs from 'yargs'
import yaml from 'js-yaml'

import { initDB } from './db'
import { startWithConfig } from './checker'
import * as queue from './queue'
import { normalizeNotifications } from './notifiers'

const readFile = util.promisify(fs.readFile)

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
		})
		.argv

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
		console.error(`'services' property must be provided at the top-level, as a dictionary`)
		yargs.showHelp()
		return
	}

	let hasErrors = false
	for (const name in config.services) {
		if (config.services.hasOwnProperty(name)) {
			if (Array.isArray(config.services[name])) {
				// for (const check of config.services[name]) {
				for (let index = 0; index < config.services[name].length; ++index) {
					const check = config.services[name][index]

					if (typeof check.name !== 'string' || !check.name) {
						console.error(`Error: 'services.${name}[${index}].name' must be a valid string`)
						hasErrors = true
					} else if (typeof check.check !== 'string' || !check.check) {
						console.error(`Error: 'services.${name}[${index}].check' must be a valid string`)
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
						console.error(`Error: 'services.${name}[${index}].interval' is not a valid time interval`)
						hasErrors = true
					}

					// Normalize image, leaving validation up to docker
					// Defaulting to custom image
					if (!check.image) {
						check.image = 'patrol-tools'
					} else if (typeof check.image !== 'string') {
						console.error(`Error: 'services.${name}[${index}].image' must be a valid docker image`)
						hasErrors = true
					}
				}
			} else {
				console.error(`Error: 'services.${name}' should be an array of checks`)
				hasErrors = true
			}
		}
	}

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
	}

	if (hasErrors) {
		console.error()
		yargs.showHelp()
		return
	}

	// Initialize collections
	initDB(config.dbDirectory)

	// Start the first scan
	await queue.Enqueue(() => startWithConfig({
		config,
	}))

	// Run the worker
	await queue.PerformWork({
		concurrency: argv.concurrency,
	})
}

main().catch(error => {
	console.error(error.stack)
	process.exit(1)
})
