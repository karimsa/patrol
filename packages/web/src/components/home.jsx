import React, { useMemo } from 'react'
import moment from 'moment'

import { Checks } from '../models/checks'
import { Config } from '../models/config'
import { useAsync } from '../state'
import { ServiceStatus } from './service-status'

export function Home() {
	const checksState = Checks.getAll()
	const configState = useAsync(Config.get)
	const [numSystemsUnhealthy, lastUpdated] = useMemo(() => {
		let numSystemsUnhealthy = 0
		let lastUpdated = -Infinity

		if (checksState.result) {
			for (const service in checksState.result) {
				if (checksState.result.hasOwnProperty(service)) {
					for (const check of checksState.result[service]) {
						lastUpdated = Math.max(lastUpdated, check.createdAt)
						if (check.serviceStatus === 'unhealthy') {
							numSystemsUnhealthy++
						}
					}
				}
			}
		}

		return [numSystemsUnhealthy, lastUpdated]
	}, [checksState.result])

	if (configState.result) {
		document.title = configState.result.title
	}

	return (
		<>
			{checksState.result && numSystemsUnhealthy === 0 && (
				<div className="bg-dark py-5">
					<div className="container">
						{configState.result && (
							<div className="row">
								<div className="col">
									<h4 className="text-white mb-4">
										{String(configState.result.title)}
									</h4>
								</div>
							</div>
						)}

						<div className="row">
							<div className="col">
								<div className="card border-none rounded overflow-hidden">
									<div className="card-body bg-success text-white d-flex justify-content-between align-items-center">
										<p className="lead mb-0 font-weight-bold">
											All Systems Operational
										</p>
										<p className="small d-none d-sm-inline-block mb-0">
											Last updated: {moment(lastUpdated).fromNow()}
										</p>
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			)}
			{checksState.result && numSystemsUnhealthy > 0 && (
				<div className="bg-dark p-5">
					<div className="container">
						<div className="row">
							<div className="col">
								<div className="card border-none rounded overflow-hidden">
									<div className="card-body bg-danger text-white d-flex justify-content-between align-items-center">
										<p className="lead mb-0 font-weight-bold">
											{numSystemsUnhealthy} Systems Are Down
										</p>
										<p className="small d-inline-block mb-0">
											Last updated: {moment(lastUpdated).fromNow()}
										</p>
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			)}

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
