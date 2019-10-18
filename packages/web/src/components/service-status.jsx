import React from 'react'

import { ServiceCheckCard } from './service-check-card'

export function ServiceStatus({ service, checks }) {
	return (
		<>
			<h4 className="font-weight-bold mb-4">{service}</h4>
			{checks.map(check => (
				<div className="row mb-4" key={check._id}>
					<div className="col">
						<ServiceCheckCard service={service} check={check} />
					</div>
				</div>
			))}
		</>
	)
}
