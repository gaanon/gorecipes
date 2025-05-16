export interface Recipe {
	id: string;
	name: string;
	ingredients: string[];
	method: string;
	photo_filename?: string; // Optional, as it might be placeholder
	created_at: string; // ISO date string
	updated_at: string; // ISO date string
}

export interface PaginatedRecipesResponse {
	recipes: Recipe[];
	total_recipes: number;
	page: number;
	limit: number;
	total_pages: number;
}