import * as os from 'os'

import Heap from 'heap'

const numCPUs = os.cpus().length
const tasks = new Heap(function(taskA, taskB) {
	return taskA.readyAt - taskB.readyAt
})
const sleep = ms => new Promise(resolve => setTimeout(resolve, ms))

export function Enqueue(task) {
	if (typeof task === 'function') {
		Enqueue({
			readyAt: Date.now(),
			run: task,
		})
		return
	}

	if (!task.readyAt) {
		task.readyAt = Date.now()
	}

	tasks.push(task)
}

export function PerformWork({ concurrency = numCPUs }) {
	return Promise.all(
		[...new Array(concurrency)].map(async () => {
			while (true) {
				const task = tasks.pop()
				if (!task) {
					await sleep(100)
					continue
				}
				if (task.readyAt && task.readyAt > Date.now()) {
					tasks.push(task)
					await sleep(100)
					continue
				}

				await task.run()
			}
		}),
	)
}
