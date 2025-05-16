import type { LoadEvent } from '@sveltejs/kit';
import type { Recipe } from '$lib/types';
import { error as svelteKitError } from '@sveltejs/kit'; // For throwing SvelteKit errors

export const load = async (event: LoadEvent) => {
	const { fetch, params } = event;
	const id = params.id;

	if (!id) {
		throw svelteKitError(400, 'Recipe ID is required for editing.');
	}

	try {
		const response = await fetch(`http://localhost:8080/api/v1/recipes/${id}`);

		if (!response.ok) {
			const errorText = await response.text();
			console.error(`Failed to fetch recipe ${id} for editing:`, response.status, errorText);
			throw svelteKitError(response.status, `Failed to load recipe for editing: ${errorText}`);
		}

		const recipe: Recipe = await response.json();
		return {
			recipe: recipe, // This will be available as 'data.recipe' in the +page.svelte
			pageError: null // Explicitly null if no error from this load function itself
		};
	} catch (e: any) {
		console.error(`Error fetching recipe ${id} for editing:`, e);
		// If e is already a SvelteKit error, rethrow it, otherwise wrap it
		if (e.status) { // Heuristic for SvelteKit error
			throw e;
		}
		throw svelteKitError(500, e.message || `An unknown error occurred while fetching recipe ${id} for editing.`);
	}
};