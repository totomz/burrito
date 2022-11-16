import { useAuth0 } from "@auth0/auth0-react";
import { SimpleTab } from "../components/card";

export function Home() {

	const {user, isAuthenticated} = useAuth0();

	return (
		<div>
			<h2>Home</h2>
			<p>{isAuthenticated}</p>
			<p>Ciao {user?.email}</p>
			<SimpleTab></SimpleTab>
		</div>
	);
}