const Notifiers = {
	webhook: {
		normalizeNotification(notification) {
			if (typeof notification.url !== 'string') {
				return `Missing valid string for 'url' attribute`
			}

			notification.method = notification.method || 'get'
			if (typeof notification.method !== 'string') {
				return `Unexpected value provided for method '${notification.method}'`
			}
		},
	},
}

export function normalizeNotification(notification) {
	if (typeof notification.type !== 'string' || !Notifiers[notification.type]) {
		return `Unrecognized 'type' attribute (must be one of: ${Object.keys(Notifiers)})`
	}
	const notifier = notification.notifier = Notifiers[notification.type]
	return notifier.normalizeNotification(notification)
}
