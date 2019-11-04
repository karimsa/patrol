import io from 'socket.io-client'

import { apiPort } from '../models/axios'
import { Checks } from '../models/checks'
// import { store } from './store'

const socket = io(
	process.env.NODE_ENV === 'production'
		? `ws://${location.host}/`
		: `ws://${location.hostname}:${apiPort}/`,
	{
		transports: ['websocket'],
	},
)

socket.on('historyUpdate', ({ service, check }) => {
	Checks.invalidateHistory({
		service,
		check,
	})
})
