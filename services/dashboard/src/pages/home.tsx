import { useAuth0 } from "@auth0/auth0-react";
import { LoginButton } from "./login";

export function Home() {

	const {user, isAuthenticated, isLoading} = useAuth0();

	if (isLoading) {
		return <div>Loading ...</div>;
	}
	
	if(!isAuthenticated) {
		return (
			<>
				<LoginButton></LoginButton>
			</>
		);
	}

	return (
		<div>
			<h2>Home</h2>
			<p>{isAuthenticated}</p>
			<p>Ciao {user?.email}</p>
		</div>
	);
}