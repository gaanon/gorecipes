import type { Load, LoadEvent } from '@sveltejs/kit'; // Use Load and LoadEvent
import type { Recipe, PaginatedRecipesResponse } from '$lib/types';

const DEFAULT_LIMIT = 25;

export const load: Load = async (event: LoadEvent) => {
const { fetch, url } = event;
const tags = url.searchParams.get('tags');
const page = parseInt(url.searchParams.get('page') || '1', 10);
const limit = parseInt(url.searchParams.get('limit') || DEFAULT_LIMIT.toString(), 10);

const queryParams = new URLSearchParams();
if (tags) {
	queryParams.append('tags', tags);
}
queryParams.append('page', page.toString());
queryParams.append('limit', limit.toString());

const apiUrl = `/api/v1/recipes?${queryParams.toString()}`;

try {
	const response = await fetch(apiUrl);

	if (!response.ok) {
		const errorText = await response.text();
		console.error('Failed to fetch recipes:', response.status, errorText);
		// Return a structure that matches what the page expects, even in error
		return {
			recipes: [],
			total_recipes: 0,
			page: page,
			limit: limit,
			total_pages: 0,
			error: `Failed to load recipes. Status: ${response.status}. ${errorText}`,
			currentTags: tags || '',
		};
	}

	const paginatedResponse: PaginatedRecipesResponse = await response.json();
	return {
		...paginatedResponse, // Spreads recipes, total_recipes, page, limit, total_pages
		error: null,
		currentTags: tags || '',
	};
} catch (e: any) {
	console.error(`Error fetching recipes (tags: ${tags}, page: ${page}, limit: ${limit}):`, e);
	return {
		recipes: [],
		total_recipes: 0,
		page: page,
		limit: limit,
		total_pages: 0,
		error: e.message || 'An unknown error occurred while fetching recipes.',
		currentTags: tags || '',
	};
}
};