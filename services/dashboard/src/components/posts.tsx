import React, { useEffect, useState } from 'react';
import { useAuth0 } from '@auth0/auth0-react';

const Posts = () => {
	const {getAccessTokenSilently} = useAuth0();
	const [posts, setPosts] = useState<any[]>([]);

	useEffect(() => {
		(async () => {
			try {
				const token = await getAccessTokenSilently({
					audience: 'https://burrito-template.daje', 
					scope: 'superuser profile email openid', 
				});
				
				const response = await fetch('http://localhost:8443/hello', {
					headers: { Authorization: `Bearer ${token}` },
				});
				setPosts(await response.json());
			}
			catch (e) {
				console.error(e);
			}
		})();
	}, [getAccessTokenSilently]);

	if (!posts) {
		return <div>Loading...</div>;
	}
		
	return (
		<ul role="list" className="space-y-3">
			{posts.map((item, index) => (
				<li key={item.id} className="overflow-hidden rounded-md bg-white px-6 py-4 shadow">
					<div className="border-b border-gray-200 bg-white px-4 py-5 sm:px-6">
						<h3 className="text-lg font-medium leading-6 text-gray-900">Job Postings</h3>
						<p className="mt-1 text-sm text-gray-500">
							{index}  {item}
						</p>
					</div>
				</li>
			))}
		</ul>
	);
};

export { Posts }