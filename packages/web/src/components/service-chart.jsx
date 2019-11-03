import { Chart } from 'chart.js'
import React from 'react'
import PropTypes from 'prop-types'
import moment from 'moment'

export function ServiceChart({ entries }) {
	const canvasRef = React.createRef()
	React.useEffect(() => {
		const metricUnit = entries.reduce(
			(unit, entry) => unit || entry.metricUnit,
			undefined,
		)
		const data = entries.map(entry => entry.metric)

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

	return <canvas ref={canvasRef}></canvas>
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
