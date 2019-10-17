import 'bootstrap/dist/css/bootstrap.min.css'
import 'bootstrap'
import '@babel/polyfill'
import $ from 'jquery'
import React from 'react'
import ReactDOM from 'react-dom'

import './app.scss'
import { Home } from './components/home'

if ((process.env.NODE_ENV || 'development') === 'development') {
	window.$ = window.jQuery = $
}

ReactDOM.render(<Home />, document.getElementById('app'))
