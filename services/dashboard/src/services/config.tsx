import { useAuth0 } from "@auth0/auth0-react";

export interface Auth {
	audience: string 
	scope: string
	domain: string
	clientID: string
} 

export class Config {

	public static ApiEndpoint(path: string): string {
		if (path[0] != "/") {
			path = `/${path}`;	
		} 
		return `${process.env.REACT_APP_APIENDPOINT}${path}`;
	}
	
	public static Authconf(): Auth {
		return {
			audience: process.env.REACT_APP_AUTH_AUDIENCE!,
			scope: process.env.REACT_APP_AUTH_SCOPE!,
			clientID: process.env.REACT_APP_AUTH_CLIENTID!,
			domain: process.env.REACT_APP_AUTH_DOMAIN!,
		}
	}

	// public static AuthconfObject(): AuthObject {
	// 	return {
	// 		audience: process.env.REACT_APP_AUTH_AUDIENCE!,
	// 		scope: process.env.REACT_APP_AUTH_SCOPE!,
	// 	}
	// }
	

}