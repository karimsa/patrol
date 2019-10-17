import axiosModule from 'axios'

export function data(fn) {
	return (...args) => {
		return fn(...args).then(d => {
			return d.data
		})
	}
}

export const axios = axiosModule.create({
	baseURL: `http://${location.hostname}:8080/api`,
	withCredentials: true,
})
