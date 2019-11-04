import 'bootstrap/dist/css/bootstrap.min.css'
import 'bootstrap'
import '@babel/polyfill'
import $ from 'jquery'
import React from 'react'
import ReactDOM from 'react-dom'
import { Provider as ReduxProvider } from 'react-redux'

import './app.scss'
import { Home } from './components/home'
import { store } from './redux/store'
import './redux/websocket'

if ((process.env.NODE_ENV || 'development') === 'development') {
	window.$ = window.jQuery = $
}

ReactDOM.render(
	<ReduxProvider store={store}>
		<Home />
	</ReduxProvider>,
	document.getElementById('app'),
)
