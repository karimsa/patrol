import assert from 'assert'

import { useMemo } from 'react'
import { useSelector } from 'react-redux'
import get from 'lodash/get'
import { createStore } from 'redux'

const idleState = result => ({
	status: 'idle',
	result,
})
const inprogressState = result => ({
	status: 'inprogress',
	result,
})
const successState = result => ({
	status: 'success',
	result,
})
const errorState = (error, result) => ({
	status: 'error',
	error,
	result,
})

const defaultState = {
	checks: idleState(),
	checkHistory: {},
}

function selectOverallHistory(state) {
	return state.checkHistory
}

export function useOverallStatus() {
	const checkHistory = useSelector(selectOverallHistory)
	return useMemo(() => {
		let numUnhealthySystems = 0
		let lastUpdated = 0

		for (const history of Object.values(checkHistory)) {
			if (!history.result) {
				return idleState()
			}

			const lastResult = history.result[history.result.length - 1]
			if (lastResult.serviceStatus === 'unhealthy') {
				numUnhealthySystems += 1
			}
			lastUpdated = Math.max(lastUpdated, lastResult.createdAt)
		}

		return successState({
			numUnhealthySystems,
			overallStatus:
				lastUpdated === 0
					? 'inprogress'
					: numUnhealthySystems === 0
					? 'healthy'
					: 'unhealthy',
			lastUpdated: new Date(lastUpdated),
		})
	}, [checkHistory])
}

export const store = createStore(
	function(state = defaultState, action) {
		switch (action.type) {
			case 'FETCH_CHECKS':
				return {
					...state,
					checks: inprogressState(state.checks.result),
				}
			case 'FETCH_CHECKS_SUCCESS':
				return {
					...state,
					checks: successState(action.result),
				}
			case 'FETCH_CHECKS_ERROR':
				return {
					...state,
					checks: errorState(action.error, state.checks.result),
				}

			case 'INVALIDATE_FETCH_CHECK_HISTORY':
				assert(action.service, `Service is required for fetching history`)
				assert(action.check, `Check is required for fetching history`)
				return {
					...state,
					checkHistory: {
						...state.checkHistory,
						[action.service + '-' + action.check]: idleState(
							get(
								state.checkHistory[action.service + '-' + action.check],
								'result',
							),
						),
					},
				}
			case 'FETCH_CHECK_HISTORY':
				assert(action.service, `Service is required for fetching history`)
				assert(action.check, `Check is required for fetching history`)
				return {
					...state,
					checkHistory: {
						...state.checkHistory,
						[action.service + '-' + action.check]: inprogressState(
							get(
								state.checkHistory[action.service + '-' + action.check],
								'result',
							),
						),
					},
				}
			case 'FETCH_CHECK_HISTORY_SUCCESS':
				assert(action.service, `Service is required for fetching history`)
				assert(action.check, `Check is required for fetching history`)
				return {
					...state,
					checkHistory: {
						...state.checkHistory,
						[action.service + '-' + action.check]: successState(action.result),
					},
				}
			case 'FETCH_CHECK_HISTORY_ERROR':
				assert(action.service, `Service is required for fetching history`)
				assert(action.check, `Check is required for fetching history`)
				return {
					...state,
					checkHistory: {
						...state.checkHistory,
						[action.service + '-' + action.check]: errorState(
							action.error,
							get(
								state.checkHistory[action.service + '-' + action.check],
								'result',
							),
						),
					},
				}

			default:
				return state
		}
	},
	window.__REDUX_DEVTOOLS_EXTENSION__ &&
		window.__REDUX_DEVTOOLS_EXTENSION__({
			trace: true,
		}),
)
