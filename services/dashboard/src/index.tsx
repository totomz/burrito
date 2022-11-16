import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import reportWebVitals from './reportWebVitals';
import { BrowserRouter } from "react-router-dom";
import { Auth0Provider } from "@auth0/auth0-react";
import { Config } from "./services/config";
import config from "tailwindcss/defaultConfig";

const root = ReactDOM.createRoot(document.getElementById('root') as HTMLElement);
root.render(
	<React.StrictMode>
		<Auth0Provider
			domain={Config.Authconf().domain}
			clientId={Config.Authconf().clientID}
			redirectUri={window.location.origin}
			cacheLocation={"localstorage"} 
			useRefreshTokens={true}
			scope={Config.Authconf().scope}	
			audience={Config.Authconf().audience}
		>
			<BrowserRouter>
				<App/>
			</BrowserRouter>
		</Auth0Provider>

	</React.StrictMode>
);

document.body.classList.add("h-full");

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
