import '@babel/polyfill'
import React from 'react'
import ReactDOMServer from 'react-dom/server'

import { App } from './app'

export const HTML = ReactDOMServer.renderToString(<App />)
