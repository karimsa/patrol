import { data, axios } from './axios'

export const Checks = {
	getAll: data(() => axios.get('/checks')),
}
