import type { LoadEvent } from '@sveltejs/kit';
import type { Recipe } from '$lib/types';

export const load = async (event: LoadEvent) => {
	const { fetch, params } = event;
	const id = params.id; // Get the 'id' from the route parameters

	try {
		const response = await fetch(`http://localhost:8080/api/v1/recipes/${id}`);

		if (!response.ok) {
			const errorText = await response.text();
			console.error(`Failed to fetch recipe ${id}:`, response.status, errorText);
			// You might want to throw an error here to trigger SvelteKit's error page
			// import { error } from '@sveltejs/kit';
			// throw error(response.status, `Failed to load recipe: ${errorText}`);
			return {
				recipe: null,
				error: `Failed to load recipe. Status: ${response.status}. ${errorText}`
			};
		}

		const recipe: Recipe = await response.json();
		return {
			recipe: recipe,
			error: null
		};
	} catch (e: any) {
		console.error(`Error fetching recipe ${id}:`, e);
		// import { error } from '@sveltejs/kit';
		// throw error(500, `Error fetching recipe: ${e.message}`);
		return {
			recipe: null,
			error: e.message || `An unknown error occurred while fetching recipe ${id}.`
		};
	}
};