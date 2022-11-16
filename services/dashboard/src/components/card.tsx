const items = [
	{id: 1},
	{id: 1},
	{id: 1},
	{id: 1},
]

const SimpleTab = () => {
	return (
		<ul role="list" className="space-y-3">
			{items.map((item) => (
				<li key={item.id} className="overflow-hidden rounded-md bg-white px-6 py-4 shadow">
					<div className="border-b border-gray-200 bg-white px-4 py-5 sm:px-6">
						<h3 className="text-lg font-medium leading-6 text-gray-900">Job Postings</h3>
						<p className="mt-1 text-sm text-gray-500">
							Lorem ipsum dolor sit amet consectetur adipisicing elit quam corrupti consectetur.
						</p>
					</div>
				</li>
			))}
		</ul>
	)
}

export { SimpleTab }