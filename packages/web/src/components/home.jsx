import React from 'react'
import moment from 'moment'
import get from 'lodash/get'

import { Checks } from '../models/checks'
import { Config } from '../models/config'
import { useAsync } from '../state'
import { ServiceStatus } from './service-status'
import { useOverallStatus } from '../redux/store'

export function Home() {
	const checksState = Checks.getAll()
	const configState = useAsync(Config.get)
	const overallState = useOverallStatus()
	const overallStatus = get(overallState.result, 'overallStatus', 'inprogress')

	if (configState.result) {
		document.title = configState.result.title
	}

	return (
		<>
			<div className="bg-dark py-5">
				<div className="container">
					<div className="row">
						<div className="col">
							<h4 className="text-white mb-4">
								{configState.result
									? String(configState.result.title)
									: 'Status'}
							</h4>
						</div>
					</div>

					<div className="row">
						<div className="col">
							<div className="card border-none rounded overflow-hidden">
								<div
									className={
										'card-body text-white d-flex justify-content-between align-items-center' +
										(overallStatus === 'healthy'
											? ' bg-success'
											: overallStatus === 'inprogress'
											? ' bg-primary'
											: ' bg-danger')
									}
								>
									<p className="lead mb-0 font-weight-bold">
										{overallStatus === 'healthy'
											? 'All Systems Operational'
											: overallStatus === 'inprogress'
											? 'Fetching service checks ...'
											: `${overallState.result.numUnhealthySystems} Systems Are Down`}
									</p>
									{overallStatus !== 'inprogress' && (
										<p className="small d-none d-sm-inline-block mb-0">
											Last updated:{' '}
											{moment(overallState.result.lastUpdated).fromNow()}
										</p>
									)}
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>

			<div className="bg-muted py-5">
				<div className="container">
					{!checksState.result && (
						<div className="row">
							<div className="col">
								<p className="lead text-center">Fetching status checks ...</p>
							</div>
						</div>
					)}
					{checksState.error && (
						<div className="row">
							<div className="col">
								<div className="alert alert-danger">
									{String(checksState.error)}
								</div>
							</div>
						</div>
					)}
					{checksState.result &&
						Object.keys(checksState.result).map((service, index) => (
							<div
								className={'row' + (index === 0 ? '' : ' mt-4')}
								key={service}
							>
								<div className="col">
									<ServiceStatus
										service={service}
										checks={checksState.result[service]}
									/>
								</div>
							</div>
						))}
				</div>
			</div>
		</>
	)
}
