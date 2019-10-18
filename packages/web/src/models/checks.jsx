import { data, axios } from './axios'

export const Checks = {
	getAll: data(() => axios.get('/checks')),
	getHistory: data(params => axios.get('/checks/history', {
		params,
	})),
}
