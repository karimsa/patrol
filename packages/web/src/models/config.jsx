import { data, axios } from './axios'

export const Config = {
	get: data(() => axios.get('/config')),
}
