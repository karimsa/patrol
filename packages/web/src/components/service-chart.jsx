import { Chart } from 'chart.js'
import React from 'react'
import PropTypes from 'prop-types'

export function ServiceChart({ data }) {
	const canvasRef = React.createRef()
	React.useEffect(() => {
		const chart = new Chart(canvasRef.current.getContext('2d'), {
			type: 'line',
			data: {
				datasets: [
					{
						backgroundColor: '#007bff',
						borderColor: '#007bff',
						data,
					},
				],
			},
			options: {
				legend: {
					display: false,
				},
				scales: {
					yAxes: [
						{
							ticks: {
								min: Math.min(...data),
								max: Math.max(...data),
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
	data: PropTypes.arrayOf(PropTypes.number.isRequired).isRequired,
}
