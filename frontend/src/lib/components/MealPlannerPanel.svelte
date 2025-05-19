<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { get } from 'svelte/store'; // Corrected import for get
	import {
		currentPlannerWeekStartDate,
		selectedPlannerDate,
		mealPlanEntriesMap,
		fetchMealPlanForWeek,
		removeRecipeFromMealPlan,
		getStartOfWeek,
		formatDateToYYYYMMDD
	} from '$lib/stores/mealPlannerStore';
	import type { MealPlanEntry } from '$lib/models/mealPlanEntry';
	import type { Recipe } from '$lib/types'; // For displaying recipe names, etc.

	// Local state for the panel
	export let isVisible = true; // Controlled by parent, e.g., main page layout

	let daysInWeek: Date[] = [];
	let recipesForSelectedDay: { entry: MealPlanEntry, recipeDetails?: Recipe }[] = []; // Recipe details to be fetched or passed

	// --- Reactive subscriptions and logic ---
	const unsubscribeCurrentWeek = currentPlannerWeekStartDate.subscribe(startDate => {
		if (startDate && typeof window !== 'undefined') { // Ensure runs in browser
			fetchMealPlanForWeek(startDate);
			generateDaysInWeek(startDate);
		}
	});

	const unsubscribeSelectedDate = selectedPlannerDate.subscribe(date => {
		if (date) {
			generateDaysInWeek(get(currentPlannerWeekStartDate)); // Re-calc days to highlight selected
			updateRecipesForSelectedDay();
		}
	});

	const unsubscribeEntriesMap = mealPlanEntriesMap.subscribe(map => {
		if (map) {
			updateRecipesForSelectedDay();
			generateDaysInWeek(get(currentPlannerWeekStartDate)); // Re-calc days to highlight days with meals
		}
	});

	function generateDaysInWeek(startDate: Date) {
		const start = getStartOfWeek(new Date(startDate));
		const newDays: Date[] = [];
		for (let i = 0; i < 7; i++) {
			const day = new Date(start);
			day.setDate(start.getDate() + i);
			newDays.push(day);
		}
		daysInWeek = newDays;
	}

	async function updateRecipesForSelectedDay() {
		const selDate = get(selectedPlannerDate);
		const entriesMap = get(mealPlanEntriesMap);
		if (!selDate || !entriesMap) {
			recipesForSelectedDay = [];
			return;
		}

		const dateStr = formatDateToYYYYMMDD(selDate);
		const entries = entriesMap.get(dateStr) || [];
		
		// Fetch recipe details for each entry
		// This is a simplified version; in a real app, you might batch these or have a recipe store
		const detailedEntries: { entry: MealPlanEntry, recipeDetails?: Recipe }[] = [];
		for (const entry of entries) {
			try {
				// TODO: Implement a more efficient way to get recipe details if not already available
				// For now, assuming we might need to fetch. This could be slow.
				// Consider a recipe store or batching API calls.
				const res = await fetch(`/api/v1/recipes/${entry.recipe_id}`);
				if (res.ok) {
					const recipe: Recipe = await res.json();
					detailedEntries.push({ entry, recipeDetails: recipe });
				} else {
					detailedEntries.push({ entry }); // Add entry even if recipe fetch fails
				}
			} catch (e) {
				console.error(`Failed to fetch recipe ${entry.recipe_id}`, e);
				detailedEntries.push({ entry });
			}
		}
		recipesForSelectedDay = detailedEntries;
	}


	onMount(() => {
		const initialDate = get(currentPlannerWeekStartDate);
		generateDaysInWeek(initialDate);
		fetchMealPlanForWeek(initialDate); // Initial fetch
		selectedPlannerDate.set(new Date()); // Set selected to today initially
	});

	onDestroy(() => {
		unsubscribeCurrentWeek();
		unsubscribeSelectedDate();
		unsubscribeEntriesMap();
	});

	// --- Event Handlers ---
	function selectDay(date: Date) {
		selectedPlannerDate.set(new Date(date));
	}

	function goToPreviousWeek() {
		const currentStart = get(currentPlannerWeekStartDate);
		const prevWeekStart = new Date(currentStart);
		prevWeekStart.setDate(currentStart.getDate() - 7);
		currentPlannerWeekStartDate.set(prevWeekStart);
		selectedPlannerDate.set(prevWeekStart); // Optionally select the first day of the new week
	}

	function goToNextWeek() {
		const currentStart = get(currentPlannerWeekStartDate);
		const nextWeekStart = new Date(currentStart);
		nextWeekStart.setDate(currentStart.getDate() + 7);
		currentPlannerWeekStartDate.set(nextWeekStart);
		selectedPlannerDate.set(nextWeekStart); // Optionally select the first day of the new week
	}

	async function handleRemoveRecipe(entryId: string) {
		if (confirm('Are you sure you want to remove this recipe from the plan?')) {
			try {
				await removeRecipeFromMealPlan(entryId);
				// The store subscription should handle UI update by refetching
			} catch (error) {
				alert(`Error removing recipe: ${error instanceof Error ? error.message : 'Unknown error'}`);
			}
		}
	}

	function handleDateInputChange(event: Event) {
		const target = event.target as HTMLInputElement;
		const dateVal = target.valueAsDate;
		if (dateVal) {
			// Adjust to local timezone if input type="date" gives UTC midnight
			const localDate = new Date(dateVal.getUTCFullYear(), dateVal.getUTCMonth(), dateVal.getUTCDate());
			currentPlannerWeekStartDate.set(getStartOfWeek(localDate));
			selectedPlannerDate.set(localDate);
		}
	}

	// --- Helper Functions ---
	function isDaySelected(date: Date): boolean {
		const selDate = get(selectedPlannerDate);
		return selDate && formatDateToYYYYMMDD(date) === formatDateToYYYYMMDD(selDate);
	}

	function hasMeals(date: Date): boolean {
		const entriesMap = get(mealPlanEntriesMap);
		return (entriesMap.get(formatDateToYYYYMMDD(date)) || []).length > 0;
	}

</script>

{#if isVisible}
<aside class="meal-planner-panel">
	<div class="panel-header">
		<h3>Weekly Meal Planner</h3>
		<button class="close-panel-button" on:click={() => isVisible = false}>&times;</button>
	</div>

	<div class="week-navigation">
		<button on:click={goToPreviousWeek}>&lt; Prev</button>
		<input 
			type="date" 
			value={formatDateToYYYYMMDD(get(selectedPlannerDate))}
			on:change={handleDateInputChange}
			aria-label="Select week by picking a date"
		/>
		<button on:click={goToNextWeek}>Next ></button>
	</div>

	<div class="calendar-view">
		{#each daysInWeek as day (formatDateToYYYYMMDD(day))}
			<button 
				class="day-cell"
				class:selected={isDaySelected(day)}
				class:has-meals={hasMeals(day)}
				on:click={() => selectDay(day)}
				title={day.toLocaleDateString(undefined, { weekday: 'long', month: 'short', day: 'numeric' })}
			>
				<span class="day-name">{day.toLocaleDateString(undefined, { weekday: 'short' })}</span>
				<span class="day-number">{day.getDate()}</span>
			</button>
		{/each}
	</div>

	<div class="selected-day-meals">
		<h4>Meals for {get(selectedPlannerDate).toLocaleDateString(undefined, { weekday: 'long', month: 'long', day: 'numeric' })}</h4>
		{#if recipesForSelectedDay.length > 0}
			<ul>
				{#each recipesForSelectedDay as item (item.entry.id)}
					<li>
						<a href="/recipes/{item.entry.recipe_id}" class="recipe-link">
							{item.recipeDetails?.name || `Recipe ID: ${item.entry.recipe_id}`}
						</a>
						<button class="remove-meal-button" on:click={() => handleRemoveRecipe(item.entry.id)} title="Remove from plan">&times;</button>
					</li>
				{/each}
			</ul>
		{:else}
			<p class="no-meals-text">No meals planned for this day.</p>
		{/if}
	</div>
</aside>
{/if}

<style>
	.meal-planner-panel {
		position: fixed;
		right: 0;
		top: 0;
		width: 350px; /* Adjust as needed */
		height: 100vh;
		background-color: var(--color-background-alt, #f9f9f9);
		border-left: 1px solid var(--color-border, #ddd);
		box-shadow: -2px 0 5px rgba(0,0,0,0.1);
		padding: 15px;
		display: flex;
		flex-direction: column;
		z-index: 1000;
		overflow-y: auto;
		transition: transform 0.3s ease-in-out;
	}
	/* Add a class to hide/show with transform if controlled by parent more smoothly */
	/* .meal-planner-panel:not(.visible) { transform: translateX(100%); } */

	.panel-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 15px;
		padding-bottom: 10px;
		border-bottom: 1px solid var(--color-border);
	}
	.panel-header h3 {
		margin: 0;
		font-size: 1.2em;
		color: var(--color-primary);
	}
	.close-panel-button {
		background: none;
		border: none;
		font-size: 1.5em;
		cursor: pointer;
		color: var(--color-text-light);
	}

	.week-navigation {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 15px;
	}
	.week-navigation button {
		padding: 5px 10px;
		background-color: var(--color-secondary);
		color: var(--color-text);
		border: none;
		border-radius: var(--border-radius-sm);
		cursor: pointer;
	}
	.week-navigation button:hover {
		background-color: var(--color-secondary-dark);
	}
	.week-navigation input[type="date"] {
		border: 1px solid var(--color-border);
		padding: 4px;
		border-radius: var(--border-radius-sm);
		font-size: 0.9em;
		max-width: 130px; /* Adjust */
	}

	.calendar-view {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		gap: 5px;
		margin-bottom: 20px;
	}
	.day-cell {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 8px 4px;
		border: 1px solid var(--color-border-light, #eee);
		border-radius: var(--border-radius-sm);
		background-color: var(--color-surface);
		cursor: pointer;
		transition: background-color 0.2s;
		font-size: 0.8em;
		min-height: 50px;
	}
	.day-cell:hover {
		background-color: var(--color-background-hover, #f0f0f0);
	}
	.day-cell.selected {
		background-color: var(--color-primary-light, #d1eaff);
		border-color: var(--color-primary);
		font-weight: bold;
	}
	.day-cell.has-meals {
		background-color: var(--color-secondary-light, #fff3cd); /* Example highlight */
	}
	.day-cell.selected.has-meals {
		background-color: var(--color-primary-light); /* Selected takes precedence or combine */
		border: 2px solid var(--color-secondary-dark);
	}
	.day-name {
		font-size: 0.9em;
		color: var(--color-text-light);
	}
	.day-number {
		font-size: 1.1em;
		font-weight: bold;
		color: var(--color-text);
	}

	.selected-day-meals {
		flex-grow: 1;
	}
	.selected-day-meals h4 {
		font-size: 1.1em;
		color: var(--color-primary-dark);
		margin-bottom: 10px;
		padding-bottom: 5px;
		border-bottom: 1px solid var(--color-border);
	}
	.selected-day-meals ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}
	.selected-day-meals li {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 8px 5px;
		border-bottom: 1px dashed var(--color-border-light, #eee);
		font-size: 0.9em;
	}
	.selected-day-meals li:last-child {
		border-bottom: none;
	}
	.recipe-link {
		color: var(--color-link);
		text-decoration: none;
		flex-grow: 1;
	}
	.recipe-link:hover {
		text-decoration: underline;
	}
	.remove-meal-button {
		background: none;
		border: none;
		color: var(--color-error);
		font-size: 1.2em;
		cursor: pointer;
		padding: 0 5px;
	}
	.no-meals-text {
		font-style: italic;
		color: var(--color-text-light);
		text-align: center;
		margin-top: 20px;
	}
</style>