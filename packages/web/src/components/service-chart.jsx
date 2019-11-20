import { Chart } from 'chart.js'
import React, { useEffect, useMemo } from 'react'
import PropTypes from 'prop-types'
import moment from 'moment'
import ms from 'ms'

function num(n) {
	return Math.round(n * 1e2) / 1e2
}

function avg(data) {
	return data.reduce((sum, item) => sum + item, 0) / data.length
}

function createLabels(entries) {
	return entries.map(entry =>
		moment(entry.createdAt).format('MMM D hh:mm:ss a'),
	)
}

const weakRefs = new WeakMap()

export function ServiceChart({ entries }) {
	const canvasRef = React.createRef()
	const metricUnit = entries.reduce(
		(unit, entry) => entry.metricUnit || unit,
		undefined,
	)
	const data = entries.map(entry => entry.metric)
	const interval =
		entries.length > 1
			? ms(
					entries.reduce((sum, _, index) => {
						if (index === 0) {
							return sum
						}
						return (
							sum + (entries[index].createdAt - entries[index - 1].createdAt)
						)
					}, 0) /
						(entries.length - 1),
			  )
			: 'unknown'

	const chartOptions = useMemo(
		() => ({
			type: 'line',
			data: {
				labels: createLabels(entries),
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
		}),
		[Date],
	)

	useEffect(() => {
		const chart = new Chart(canvasRef.current.getContext('2d'), chartOptions)
		weakRefs.set(chartOptions, chart)
		return () => chart.destroy()
	}, [Date])

	const chart = weakRefs.get(chartOptions)
	if (chart) {
		chartOptions.options.scales.yAxes[0].ticks.suggestedMin = Math.min(...data)
		chartOptions.options.scales.yAxes[0].ticks.suggestedMax = Math.max(...data)

		while (chartOptions.data.labels.length > 0) {
			chartOptions.data.labels.pop()
		}
		createLabels(entries).forEach(label => {
			chartOptions.data.labels.push(label)
		})

		while (chartOptions.data.datasets[0].data.length > 0) {
			chartOptions.data.datasets[0].data.pop()
		}
		entries.forEach(entry => {
			chartOptions.data.datasets[0].data.push(entry.metric)
		})

		chart.update()
	}

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
					<span className="mx-2">&bull;</span>
					<span className="text-primary mr-2">Interval:</span>
					<span>{interval}</span>
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
