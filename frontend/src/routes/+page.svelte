<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { onMount, afterUpdate } from 'svelte'; // Added afterUpdate
	import type { PageData } from './$types';
	import type { Recipe, PaginatedRecipesResponse } from '$lib/types';
	import RecipeCard from '$lib/components/RecipeCard.svelte';
	import MealPlannerPanel from '$lib/components/MealPlannerPanel.svelte'; // Added

	export let data: PageData;

	// Reactive state for all displayed recipes
	let displayedRecipes: Recipe[] = [];
	let currentPage: number = 1;
	let isLoadingMore = false;
	let filterTagsInput = data.currentTags || '';
	let isPlannerVisible = false; // Added for planner visibility

	// Initialize displayedRecipes when data first loads or changes significantly (e.g., filter applied)
	$: {
		if (data.recipes && data.page === 1) {
			// Reset if it's the first page (e.g., new filter or initial load)
			displayedRecipes = data.recipes;
			currentPage = data.page;
		} else if (data.recipes && data.page > currentPage) {
			// Append if it's a subsequent page from "Load More"
			displayedRecipes = [...displayedRecipes, ...data.recipes];
			currentPage = data.page;
		} else if (data.recipes) {
			// Fallback or if page number hasn't changed but data might have (e.g. invalidateAll)
			// This might need refinement based on specific scenarios.
			// For now, if page is not 1 and not greater, assume it's a refresh of current data set.
			displayedRecipes = data.recipes;
			currentPage = data.page;
		}
		// Removed problematic block that reset filterTagsInput
		// The initial declaration `let filterTagsInput = data.currentTags || '';`
		// and `bind:value` on the input should handle synchronization correctly
		// when `data` (and thus `data.currentTags`) is updated from the load function.
	}


	function applyFilter() {
		const params = new URLSearchParams();
		if (filterTagsInput.trim() !== '') {
			params.set('tags', filterTagsInput.trim());
		}
		// When applying a new filter, always go to page 1
		params.set('page', '1');
		goto(`/?${params.toString()}`);
	}

	function clearFilter() {
		filterTagsInput = '';
		// When clearing filter, always go to page 1
		goto(`/?page=1`);
	}

	async function loadMoreRecipes() {
		if (isLoadingMore || !data.total_pages || currentPage >= data.total_pages) {
			return;
		}
		isLoadingMore = true;

		const nextPage = currentPage + 1;
		const params = new URLSearchParams(window.location.search); // Preserve existing filters
		params.set('page', nextPage.toString());
		
		// We use goto to trigger the load function in +page.ts
		// The reactive block above will handle appending new recipes
		await goto(`/?${params.toString()}`, { keepFocus: true, noScroll: true });
		
		isLoadingMore = false;
	}

	// $: console.log('Data from load:', data);
	// $: console.log('Displayed Recipes:', displayedRecipes.length, 'Current Page:', currentPage);

</script>

<svelte:head>
	<title>GoRecipes - All Recipes {data.currentTags ? `(Filtered by: ${data.currentTags})` : ''}</title>
</svelte:head>

<div class="main-container">
	<div class="page-header">
		{#if data.currentTags}
			<p class="filter-indicator-standalone">Filtered by: {data.currentTags}</p>
		{/if}
		<button class="button planner-toggle-button" on:click={() => isPlannerVisible = !isPlannerVisible}>
			{isPlannerVisible ? 'Hide' : 'Show'} Planner
		</button>
	</div>

	<form class="filter-form" on:submit|preventDefault={applyFilter}>
		<label for="filter-tags-input" class="filter-label">Filter by Ingredients</label>
		<div class="filter-controls">
			<input
				type="text"
				id="filter-tags-input"
				class="filter-input"
				bind:value={filterTagsInput}
				placeholder="e.g., chicken, tomato, basil"
			/>
			<button type="submit" class="button filter-action-button">
				<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="18" height="18" style="margin-right: 6px;">
					<path fill-rule="evenodd" d="M9 3.5a5.5 5.5 0 100 11 5.5 5.5 0 000-11zM2 9a7 7 0 1112.452 4.391l3.328 3.329a.75.75 0 11-1.06 1.06l-3.329-3.328A7 7 0 012 9z" clip-rule="evenodd" />
				</svg>
				Filter
			</button>
			{#if data.currentTags || filterTagsInput}
				<button type="button" on:click={clearFilter} class="button secondary clear-filter-button">
					<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="18" height="18" style="margin-right: 6px;">
						<path fill-rule="evenodd" d="M2.513 2.513C2.826 2.2 3.276 2 3.75 2H16.25c.474 0 .924.2 1.237.513l-2.612 2.612L12.05 8.05l-2.225 2.224a.75.75 0 001.06 1.06L13.11 9.11l2.829 2.828-2.414 2.415a.75.75 0 01-1.06 0L10.24 12.13l-2.47 2.47a.75.75 0 01-1.06 0L4.485 12.375l-1.972 1.972C2.2 14.037 2 13.587 2 13.113V3.75c0-.474.2-.924.513-1.237zm14.193 11.099L6.53 3.436A.75.75 0 017.59 3.5l8.114 8.113a.75.75 0 010 1.06l-2.087 2.087.075.075a.75.75 0 001.06-1.06l-.074-.075z" clip-rule="evenodd" />
					</svg>
					Clear
				</button>
			{/if}
		</div>
	</form>

	{#if data.error}
		<p class="info-message error-state">{data.error}</p>
	{:else if displayedRecipes.length > 0}
		<div class="recipes-grid">
			{#each displayedRecipes as recipe (recipe.id)}
				<RecipeCard {recipe} />
			{/each}
		</div>
		{#if data.total_pages && currentPage < data.total_pages}
			<div class="load-more-container">
				<button on:click={loadMoreRecipes} disabled={isLoadingMore} class="button primary load-more-button">
					{#if isLoadingMore}
						<span class="spinner"></span> Loading...
					{:else}
						Load More Recipes ({displayedRecipes.length} / {data.total_recipes})
					{/if}
				</button>
			</div>
		{/if}
	{:else if data.currentTags}
		<p class="info-message">No recipes found for "{data.currentTags}". Try different ingredients or clear the filter.</p>
	{:else}
		<p class="info-message">No recipes yet! Be the first to <a href="/recipes/new">add one</a>.</p>
	{/if}
</div>

<MealPlannerPanel bind:isVisible={isPlannerVisible} />

<style>
	/* .page-container is now .main-container from app.css */

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 30px; /* Increased margin */
		flex-wrap: wrap;
		gap: 15px; /* Gap for wrapping */
	}

	/* .page-title styles removed */
	.filter-indicator-standalone { /* New style for standalone filter indicator */
		font-size: 1.1em; /* Make it a bit more prominent if standalone */
		font-weight: 500;
		color: var(--color-text-light);
		margin-bottom: 15px; /* Add some space below if it's the only thing in header */
		/* text-align: center; */ /* Adjusted for new button */
		/* width: 100%; */ /* Adjusted for new button */
	}
	.planner-toggle-button {
		/* Basic styling, can be improved */
		background-color: var(--color-accent, #ff9800);
		color: white;
		padding: 8px 15px;
		font-size: 0.9em;
	}
	.planner-toggle-button:hover {
		background-color: #e68a00; /* Darker accent */
	}


	/* .create-recipe-button styles removed as the button is gone */

	.filter-form {
		margin-bottom: 40px; /* Increased margin */
		padding: 20px;
		background-color: var(--color-surface); /* Use surface color for forms */
		border-radius: var(--border-radius);
		box-shadow: var(--shadow-sm);
	}
	.filter-label {
		display: block;
		margin-bottom: 12px;
		font-weight: 600; /* Bolder label */
		font-size: 1.1em;
		color: var(--color-text-light);
	}
	.filter-controls {
		display: flex;
		gap: 12px; /* Slightly increased gap */
		align-items: center;
		flex-wrap: wrap;
	}
	.filter-input {
		/* Uses global input styles from app.css */
		flex-grow: 1;
		min-width: 250px; /* Increased min-width */
	}

	.button { /* General button class if not fully covered by app.css */
		border: none;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}
	.button.secondary { /* For clear button */
		background-color: var(--color-text-light);
		color: white;
	}
	.button.secondary:hover {
		background-color: var(--color-text);
	}

	.filter-action-button {
		background-color: var(--color-secondary); /* Using secondary for filter button */
		color: var(--color-text); /* Dark text on yellow */
	}
	.filter-action-button:hover {
		background-color: #ffca28; /* Slightly darker yellow */
		box-shadow: var(--shadow-sm);
	}
	.clear-filter-button {
		/* Uses .button.secondary */
	}


	.recipes-grid {
		display: grid;
		/* Adjusted minmax for potentially larger cards or different screen sizes */
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 25px; /* Increased gap */
	}

	.info-message {
		text-align: center;
		font-size: 1.1em;
		color: var(--color-text-light);
		padding: 20px;
		background-color: var(--color-surface);
		border-radius: var(--border-radius);
		box-shadow: var(--shadow-sm);
	}
	.info-message.error-state {
		color: var(--color-error);
		background-color: #ffebee; /* Light red background for errors */
		border: 1px solid var(--color-error);
	}
	.info-message a {
		font-weight: 600;
	}

	.load-more-container {
		text-align: center;
		margin-top: 30px;
		margin-bottom: 20px;
	}
	.load-more-button {
		padding: 12px 25px;
		font-size: 1.1em;
	}
	.load-more-button .spinner {
		width: 18px;
		height: 18px;
		border: 3px solid rgba(255,255,255,0.3);
		border-radius: 50%;
		border-top-color: white;
		animation: spin 1s ease-infinite;
		margin-right: 10px;
		display: inline-block; /* Ensure spinner aligns well */
		vertical-align: middle;
	}
	@keyframes spin {
		to { transform: rotate(360deg); }
	}

</style>
