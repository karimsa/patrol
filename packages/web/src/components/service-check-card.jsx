/** @jsx jsx */

import PropTypes from 'prop-types'
import { jsx, css } from '@emotion/core'
import moment from 'moment'

import { useAsync } from '../state'
import { Checks, CheckType } from '../models/checks'

const numHistoryBars = 80
const barWidth = 2
const barSpacing = 2
const svgWidth = numHistoryBars * barWidth + (numHistoryBars - 1) * barSpacing

const colorGray = '#d9dbde'
const colorGreen = '#00eb8b'
const colorYellow = '#ffbc62'
const colorBlue = '#007bff'

function createElms(length, fn) {
	const elms = new Array(length)
	for (let i = 0; i < length; ++i) {
		elms[i] = fn(i)
	}
	return elms
}

export function ServiceCheckCard({ service, check }) {
	const historyState = useAsync(
		() =>
			Checks.getHistory({
				service,
				check: check.check,
			}),
		[service, check],
	)
	const numDimBars = historyState.result
		? numHistoryBars - historyState.result.length
		: 0

	return (
		<div className="card">
			<div className="card-body">
				<div className="row">
					<div className="col">
						<div className="d-flex justify-content-between">
							<p className="font-weight-bold mb-0 d-inline-block">
								{check.check}
							</p>
							<p
								className={
									'font-weight-bold mb-0 d-inline-block d-flex align-items-center ' +
									(check.serviceStatus === 'healthy'
										? 'text-success'
										: check.serviceStatus === 'unhealthy'
										? 'text-warn'
										: 'text-primary')
								}
							>
								<span>
									{check.serviceStatus === 'healthy'
										? 'Healthy'
										: check.serviceStatus === 'unhealthy'
										? 'Unhealthy'
										: 'In Progress'}
								</span>
								<span className="small text-muted ml-2">
									{moment(check.createdAt).fromNow()}
								</span>
							</p>
						</div>
					</div>
				</div>

				<div className="row mt-4">
					<div className="col">
						{historyState.error && (
							<div className="alert alert-danger">
								{String(historyState.error)}
							</div>
						)}
						{historyState.status === 'inprogress' && (
							<p className="mb-0 text-muted">Fetching service history ...</p>
						)}
						{historyState.result && (
							<svg
								className="w-100"
								viewBox={`0 0 ${svgWidth} 10`}
								css={css`
									height: 2rem;
								`}
							>
								{createElms(numDimBars, index => (
									<rect
										key={index}
										height="10"
										width={barWidth}
										x={index * (barWidth + barSpacing)}
										y="0"
										fill={colorGray}
									/>
								))}

								{historyState.result.map((historyEntry, index) => (
									<rect
										key={index + numDimBars}
										height="10"
										width={barWidth}
										x={(index + numDimBars) * (barWidth + barSpacing)}
										y="0"
										fill={
											historyEntry.serviceStatus === 'healthy'
												? colorGreen
												: historyEntry.serviceStatus === 'unhealthy'
												? colorYellow
												: colorBlue
										}
									/>
								))}
							</svg>
						)}
					</div>
				</div>
			</div>
		</div>
	)
}

ServiceCheckCard.propTypes = {
	service: PropTypes.string.isRequired,
	check: CheckType.isRequired,
}
