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
			<h1>Basic Example</h1>

			<p>
				This example demonstrates some of the core features of React Router
				including nested <code>&lt;Route&gt;</code>s,{" "}
				<code>&lt;Outlet&gt;</code>s, <code>&lt;Link&gt;</code>s, and using a
				"*" route (aka "splat route") to render a "not found" page when someone
				visits an unrecognized URL.
			</p>

			{/* Routes nest inside one another. Nested route paths build upon
            parent route paths, and nested route elements render inside
            parent route elements. See the note about <Outlet> below. */}
			<Routes>
				<Route path="/" element={<Layout/>}>
					<Route index element={<Home/>}/>
					<Route path="about" element={<About/>}/>
					
					{/* Using path="*"" means "match anything", so this route
                acts like a catch-all for URLs that we don't have explicit
                routes for. */}
					<Route path="*" element={<NotFound/>}/>
				</Route>
			</Routes>
		</div>
	);
}

export default App;
