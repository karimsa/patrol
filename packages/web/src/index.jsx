import 'bootstrap/dist/css/bootstrap.min.css'
import 'bootstrap'
import '@babel/polyfill'
import $ from 'jquery'
import React from 'react'
import ReactDOM from 'react-dom'

import './app.scss'
import './redux/websocket'
import { App } from './app'

if ((process.env.NODE_ENV || 'development') === 'development') {
	window.$ = window.jQuery = $
}

const root = document.getElementById('app')
if (root.innerText.length) {
	ReactDOM.hydrate(<App />, root)
} else {
	ReactDOM.render(<App />, root)
}
