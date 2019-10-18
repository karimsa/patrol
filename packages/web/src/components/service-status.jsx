import React from 'react'
// import { jsx, css } from '@emotion/core'
import moment from 'moment'
import ms from 'ms'

// const numHistoryBars = 80
// const barWidth = 2
// const barSpacing = 2
// const svgWidth = (numHistoryBars * barWidth) + ((numHistoryBars - 1) * barSpacing)
// const historyList = [... new Array(numHistoryBars)]

export function ServiceStatus({ service, checks }) {
	return (
		<>
			<h4 className="font-weight-bold mb-4">{service}</h4>
			{checks.map(check => (
				<div className="card" key={check._id}>
					<div className="card-body">
						<div className="row">
							<div className="col">
								<div className="d-flex justify-content-between">
									<p className="font-weight-bold mb-0 d-inline-block">{check.check}</p>
									<p className={"font-weight-bold mb-0 d-inline-block d-flex align-items-center " + (check.serviceStatus === 'healthy' ? 'text-success' : 'text-warn')}>
										<span>{check.serviceStatus === 'healthy' ? 'Healthy' : 'Unhealthy'}</span>
										<span className="small text-muted ml-2">{moment(check.createdAt).fromNow()}</span>
									</p>
								</div>
							</div>
						</div>

						{/* TODO: Add history */}
						{/* <div className="row mt-4">
							<div className="col">
								<svg className="w-100" viewBox={`0 0 ${svgWidth} 10`} css={css`height: 2rem`}>
									{historyList.map((_, index) => (
										<rect
											key={index}
											height="10"
											width={barWidth}
											x={index * (barWidth + barSpacing)}
											y="0"
											fill="#00eb8b"
										/>
									))}
								</svg>
							</div>
						</div> */}
					</div>
				</div>
			))}
		</>
	)
}
