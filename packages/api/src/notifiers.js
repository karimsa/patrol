import request from 'request-promise-native'
import Mustache from 'mustache'
import { logger } from '@karimsa/boa'

import * as queue from './queue'

const Notifiers = {
	webhook: {
		normalizeNotification(notification) {
			request.defaults(notification.options)
		},

		async sendNotification(notification, serviceCheck) {
			const body = notification.options.body ? Mustache.render(notification.options.body, serviceCheck) : undefined
			const requestOptions = {
				...notification.options,
				body,
			}

			logger.debug('patrol', `Executing outgoing webhook with options: %O`, requestOptions)
			await request(requestOptions)
		},
	},
}

function normalizeNotification(notification) {
	if (typeof notification.type !== 'string' || !Notifiers[notification.type]) {
		return `Unrecognized 'type' attribute (must be one of: ${Object.keys(Notifiers)})`
	}
	const notifier = notification.notifier = Notifiers[notification.type]
	return notifier.normalizeNotification(notification)
}

export function normalizeNotifications(notifications) {
	if (!Array.isArray(notifications)) {
		return `notifications must be an array`
	}

	let errorBuffer = ''
	for (let index = 0; index < notifications.length; ++index) {
		const error = normalizeNotification(notifications[index])
		if (error) {
			errorBuffer += `notifications[${index}]: ${error}\n`
		}
	}
	return errorBuffer
}

export function sendNotifications(notifications, serviceCheck) {
	if (notifications) {
		for (const notification of notifications) {
			queue.Enqueue(() => {
				logger.info(`Sending notification of type %O for check %O in service %O`, notification.type, serviceCheck.check.name, serviceCheck.service)
				return notification.notifier.sendNotification(notification, serviceCheck)
			})
		}
	}
}
