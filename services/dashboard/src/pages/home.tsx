import { useAuth0 } from "@auth0/auth0-react";

export function Home() {

	const {user, isAuthenticated} = useAuth0();

	return (
		<div>
			<h2>Home</h2>
			<p>{isAuthenticated}</p>
			<p>Ciao {user?.email}</p>
		</div>
	);
}