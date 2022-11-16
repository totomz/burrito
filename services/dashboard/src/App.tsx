import React from 'react';
import './App.css';
import { Routes, Route } from "react-router-dom";
import { Layout } from "./components/layout";
import { Home } from "./pages/home";
import { About } from "./pages/about";
import { NotFound } from "./pages/404";
import { useAuth0 } from "@auth0/auth0-react";
import { Login  } from "./pages/login";

function App() {
	const { isAuthenticated, isLoading } = useAuth0();
	
	if (isLoading) {
		return <div>Loading ...</div>;
	}
	
	if(!isAuthenticated) {
		return (
			<Login></Login>
		);
	}
	
	return (
		<div>
			<Routes>
				<Route path="/" element={<Layout/>}>
					<Route index element={<Home/>}/>
					<Route path="about" element={<About/>}/>
					
					<Route path="*" element={<NotFound/>}/>
				</Route>
			</Routes>
		</div>
	);
}

export default App;
