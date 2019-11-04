import PropTypes from 'prop-types'
import { useSelector } from 'react-redux'

import { data, axios } from './axios'
import { store } from '../redux/store'

function createAsyncAction(typePrefix, fn) {
	return function(args) {
		store.dispatch({ type: typePrefix, ...args })
		Promise.resolve(fn(args))
			.then(result =>
				store.dispatch({ type: `${typePrefix}_SUCCESS`, result, ...args }),
			)
			.catch(error =>
				store.dispatch({
					type: `${typePrefix}_ERROR`,
					error,
					debug: String(error.stack),
					...args,
				}),
			)
	}
}

function createInvalidationAction(typeSuffix) {
	return (props = {}) =>
		store.dispatch({ type: `INVALIDATE_${typeSuffix}`, ...props })
}

const fetch = {
	getAll: createAsyncAction('FETCH_CHECKS', data(() => axios.get('/checks'))),
	getHistory: createAsyncAction(
		'FETCH_CHECK_HISTORY',
		data(params =>
			axios.get('/checks/history', {
				params,
			}),
		),
	),
}

export const Checks = {
	getAll() {
		const checksState = useSelector(state => state.checks)
		if (checksState.status === 'idle') {
			fetch.getAll()
		}
		return checksState
	},

	invalidateHistory: createInvalidationAction('FETCH_CHECK_HISTORY'),

	getHistory(params) {
		const checkHistoryState = useSelector(
			state =>
				state.checkHistory[params.service + '-' + params.check] || {
					status: 'idle',
				},
		)
		if (checkHistoryState.status === 'idle') {
			fetch.getHistory(params)
		}
		return checkHistoryState
	},
}

export const CheckType = PropTypes.shape({
	_id: PropTypes.string.isRequired,
	check: PropTypes.string.isRequired,
	serviceStatus: PropTypes.string.isRequired,
	createdAt: PropTypes.number.isRequired,
	output: PropTypes.string.isRequired,
})
