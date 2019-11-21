import axiosModule from 'axios'

export function data(fn) {
	return (...args) => {
		return fn(...args).then(d => {
			return d.data
		})
	}
}

export const apiPort = 8081

export const axios = axiosModule.create({
	baseURL: process.env.IS_SERVER
		? 'http://localhost:1/'
		: process.env.NODE_ENV === 'production'
		? `${location.protocol}//${location.host}/api`
		: `http://${location.hostname}:${apiPort}/api`,
	withCredentials: true,
})
