import { Chart } from 'chart.js'
import React from 'react'
import PropTypes from 'prop-types'
import moment from 'moment'

function num(n) {
	return Math.round(n * 1e2) / 1e2
}

function avg(data) {
	return data.reduce((sum, item) => sum + item, 0) / data.length
}

export function ServiceChart({ entries }) {
	const canvasRef = React.createRef()
	const metricUnit = entries.reduce(
		(unit, entry) => unit || entry.metricUnit,
		undefined,
	)
	const data = entries.map(entry => entry.metric)

	React.useEffect(() => {
		const chart = new Chart(canvasRef.current.getContext('2d'), {
			type: 'line',
			data: {
				labels: entries.map(entry =>
					moment(entry.createdAt).format('MMM D hh:mm:ss a'),
				),
				datasets: [
					{
						backgroundColor: '#007bff',
						borderColor: '#007bff',
						data,
						fill: false,
					},
				],
			},
			options: {
				legend: {
					display: false,
				},
				scales: {
					xAxes: [
						{
							ticks: {
								display: false,
							},
						},
					],
					yAxes: [
						{
							scaleLabel: {
								display: Boolean(metricUnit),
								labelString: metricUnit,
							},
							ticks: {
								display: false,
								suggestedMin: Math.min(...data),
								suggestedMax: Math.max(...data),
							},
						},
					],
				},
			},
		})
		return () => chart.destroy()
	}, [canvasRef])

	return (
		<React.Fragment>
			<canvas ref={canvasRef}></canvas>
			<div className="mt-4">
				<p className="mb-0 text-center">
					<span className="text-primary mr-2">Average:</span>
					<span>
						{num(avg(data))}
						{metricUnit && ' ' + metricUnit}
					</span>
					<span className="mx-2">&bull;</span>
					<span className="text-primary mr-2">Min:</span>
					<span>
						{num(Math.min(...data))}
						{metricUnit && ' ' + metricUnit}
					</span>
					<span className="mx-2">&bull;</span>
					<span className="text-primary mr-2">Max:</span>
					<span>
						{num(Math.max(...data))}
						{metricUnit && ' ' + metricUnit}
					</span>
				</p>
			</div>
		</React.Fragment>
	)
}

ServiceChart.propTypes = {
	entries: PropTypes.arrayOf(
		PropTypes.shape({
			metricUnit: PropTypes.string.isRequired,
			metric: PropTypes.number.isRequired,
			createdAt: PropTypes.number.isRequired,
		}),
	).isRequired,
}
