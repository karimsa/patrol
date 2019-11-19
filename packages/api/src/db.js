import * as path from 'path'

import Datastore from 'nedb'

const db = {}

function createModel(name, dbPath, defaultOpts = {}) {
	const store = new Datastore({
		filename: path.resolve(dbPath, name + '.db'),
		autoload: true,
	})

	// Auto compact every 30 mins, to ensure that we are not over-storing
	// data
	store.persistence.setAutocompactionInterval(1000 * 60 * 30)

	const defaultSort = defaultOpts.sort || {}

	return {
		insert(doc) {
			return new Promise((resolve, reject) => {
				store.insert(doc, err => {
					if (err) reject(err)
					else resolve()
				})
			})
		},

		update(filter, doc, opts = {}) {
			return new Promise((resolve, reject) => {
				store.update(filter, doc, opts, err => {
					if (err) reject(err)
					else resolve()
				})
			})
		},

		remove(filter, opts = {}) {
			return new Promise((resolve, reject) => {
				store.remove(filter, opts, err => {
					if (err) reject(err)
					else resolve()
				})
			})
		},

		count(filter) {
			return new Promise((resolve, reject) => {
				store.count(filter, (error, count) => {
					if (error) reject(error)
					else resolve(count)
				})
			})
		},

		find(filter, opts = {}) {
			return new Promise((resolve, reject) => {
				let cursor = store.find(filter).sort(opts.sort || defaultSort)
				if ('limit' in opts) {
					cursor = cursor.limit(opts.limit)
				}

				cursor.exec((err, results) => {
					if (err) reject(err)
					else resolve(results)
				})
			})
		},

		findOne(filter, opts = {}) {
			return this.find(filter, { ...opts, limit: 1 }).then(
				results => results[0],
			)
		},

		exists(filter) {
			return this.findOne(filter).then(res => Boolean(res))
		},
	}
}

export function initDB(dbPath) {
	Object.assign(db, {
		Checks: createModel('checks', dbPath),
	})
}

export function model(name) {
	const model = db[name]
	if (!model) {
		throw new Error(`Model with name '${name}' does not exist`)
	}
	return model
}
