export const idleState = result => ({
	status: 'idle',
	result,
})
export const inprogressState = result => ({
	status: 'inprogress',
	result,
})
export const successState = result => ({
	status: 'success',
	result,
})
export const errorState = (error, result) => ({
	status: 'error',
	error,
	result,
})
