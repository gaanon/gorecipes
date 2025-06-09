<script lang="ts">
	import type { Recipe } from '$lib/types';
	import AddToPlanModal from '$lib/components/AddToPlanModal.svelte'; // Added

	export let recipe: Recipe;
	let showAddToPlanModal = false; // Added

	const baseImageUrl = '/uploads/images/';
	$: imageUrl = recipe.photo_filename ? `${baseImageUrl}${recipe.photo_filename}` : ''; // Handle missing photo_filename gracefully

	function openAddToPlanModal(event: MouseEvent) {
		event.preventDefault(); // Prevent link navigation if button is inside <a>
		event.stopPropagation(); // Prevent event bubbling
		showAddToPlanModal = true;
	}
</script>

<div class="recipe-card-wrapper">
	<a href="/recipes/{recipe.id}" class="recipe-card-link">
		<div class="recipe-card">
			<div class="photo-container">
				{#if imageUrl}
				<img src={imageUrl} alt={recipe.name} class="recipe-photo" />
			{:else}
				<div class="recipe-photo-placeholder">
					<span>No Image</span>
				</div>
			{/if}
		</div>
		<div class="card-content">
			<h3 class="recipe-name">{recipe.name}</h3>
			<!-- Future: Maybe a short description or key ingredients -->
			<span class="view-details-button">View Recipe</span>
		</div>
	</div>
	</a>
	<button class="add-to-plan-flt-btn" on:click={openAddToPlanModal} title="Add to Meal Plan">
		<span>+</span>📅
	</button>
</div>

{#if showAddToPlanModal}
	<AddToPlanModal
		bind:showModal={showAddToPlanModal}
		recipeId={recipe.id}
		recipeName={recipe.name}
		on:close={() => showAddToPlanModal = false}
	/>
{/if}

<style>
	.recipe-card-wrapper {
		position: relative; /* For positioning the floating button */
		display: block; /* Or inline-block, depending on layout needs */
	}
	.recipe-card-wrapper .add-to-plan-flt-btn {
		position: absolute;
		top: 10px;
		right: 10px;
		z-index: 10;
		background-color: var(--color-accent, #ff9800);
		color: white;
		border: none;
		border-radius: 50%;
		width: 40px;
		height: 40px;
		font-size: 1.2em; /* Adjust icon size */
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		box-shadow: var(--shadow-md);
		opacity: 0; /* Hidden by default */
		transition: opacity 0.2s ease-in-out, transform 0.2s ease-in-out;
		transform: scale(0.8);
	}
	.recipe-card-wrapper:hover .add-to-plan-flt-btn {
		opacity: 1; /* Show on hover */
		transform: scale(1);
	}
	.add-to-plan-flt-btn span {
		position: relative;
		top: -1px; /* Fine-tune icon position */
	}


	.recipe-card-link {
		text-decoration: none; /* Remove underline from the link wrapping the card */
		color: inherit; /* Inherit text color */
		display: block; /* Make the link a block to contain the card properly */
		transition: transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out;
	}
	.recipe-card-link:hover {
		transform: translateY(-5px);
		box-shadow: var(--shadow-md); /* Enhance shadow on hover for the link itself */
	}
	.recipe-card-link:hover .recipe-card {
		/* If you want shadow on card itself, but link hover is often better for transform */
		/* box-shadow: var(--shadow-md); */
	}


	.recipe-card {
		background-color: var(--color-surface);
		border-radius: var(--border-radius);
		box-shadow: var(--shadow-sm);
		overflow: hidden; /* Ensures content respects border-radius, esp. images */
		display: flex;
		flex-direction: column;
		height: 100%; /* Make card take full height of grid cell if needed */
		/* width: 300px; /* Removed fixed width to be responsive in a grid */
		/* margin: 16px; /* Margin should be handled by the grid gap */
		transition: box-shadow 0.2s ease-in-out; /* Smooth shadow transition if card itself has hover */
	}

	.photo-container {
		width: 100%;
		height: 200px; /* Fixed height for image container */
		background-color: #e0e0e0; /* Placeholder bg if image is missing or loading */
		display: flex; /* For placeholder text centering */
		align-items: center;
		justify-content: center;
	}

	.recipe-photo {
		width: 100%;
		height: 100%;
		object-fit: cover; /* Crop image to fit, maintaining aspect ratio */
	}

	.recipe-photo-placeholder {
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		background-color: var(--color-border); /* Use a theme color */
		color: var(--color-text-light);
		font-size: 0.9em;
	}
	.recipe-photo-placeholder span {
		padding: 5px 10px;
		background-color: rgba(0,0,0,0.1);
		border-radius: 4px;
	}

	.card-content {
		padding: 15px;
		display: flex;
		flex-direction: column;
		flex-grow: 1; /* Allows content to fill space, pushing button to bottom */
		text-align: left; /* Changed from center */
	}

	.recipe-name {
		font-size: 1.25em; /* Slightly adjusted */
		font-weight: 600; /* From global styles, but can be specific */
		color: var(--color-text);
		margin-top: 0;
		margin-bottom: 10px;
		line-height: 1.3;
	}

	.view-details-button {
		display: inline-block; /* Make it behave like a button */
		margin-top: auto; /* Pushes to the bottom of the card content */
		padding: 8px 15px;
		background-color: var(--color-primary);
		color: white;
		border-radius: calc(var(--border-radius) - 4px); /* Slightly smaller radius */
		text-align: center;
		font-weight: 500;
		font-size: 0.9em;
		transition: background-color 0.2s ease;
	}

	.recipe-card-link:hover .view-details-button {
		background-color: var(--color-primary-dark);
	}

</style>
