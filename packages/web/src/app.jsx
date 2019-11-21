import React from 'react'
import { Provider as ReduxProvider } from 'react-redux'

import './app.scss'
import { Home } from './components/home'
import { store } from './redux/store'

export function App() {
	return (
		<ReduxProvider store={store}>
			<Home />
		</ReduxProvider>
	)
}
