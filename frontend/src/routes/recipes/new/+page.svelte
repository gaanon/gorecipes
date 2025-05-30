<script lang="ts">
	import { goto } from '$app/navigation';
	import { enhance } from '$app/forms';
	import type { ActionResult } from '@sveltejs/kit';
	import type { Recipe } from '$lib/types'; // For potential response typing

	let recipeName = '';
	let ingredientsStr = ''; // Comma-separated string for simplicity
	let method = '';
	let photoFile: FileList | null = null;
	let isLoading = false;
	let formError: string | null = null;
	let formSuccess: string | null = null;

	// This function will be called after the form submission (if using progressive enhancement)
	function handleResult(event: CustomEvent<ActionResult>) {
		isLoading = false;
		if (event.detail.type === 'success') {
			formSuccess = 'Recipe created successfully!';
			formError = null;
			// Optionally, redirect to the new recipe page or home
			// For now, just clear the form and show success
			const createdRecipe = event.detail.data?.recipe as Recipe; // Assuming backend returns the recipe
			if (createdRecipe && createdRecipe.id) {
				goto(`/recipes/${createdRecipe.id}`);
			} else {
				// Fallback or stay on page
				recipeName = '';
				ingredientsStr = '';
				method = '';
				const photoInput = document.getElementById('photo') as HTMLInputElement;
				if (photoInput) photoInput.value = ''; // Clear file input
			}
		} else if (event.detail.type === 'failure') {
			formError = event.detail.data?.error || 'Failed to create recipe. Please try again.';
			formSuccess = null;
		} else if (event.detail.type === 'error') {
			formError = event.detail.error.message || 'An unexpected error occurred.';
			formSuccess = null;
		}
	}
</script>

<svelte:head>
	<title>Create New Recipe - GoRecipes</title>
</svelte:head>

<div class="main-container form-page-container">
	<h1 class="page-title">Craft a New Recipe</h1>

	{#if formSuccess}
		<div class="message success-message">
			<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="20" height="20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clip-rule="evenodd" /></svg>
			<span>{formSuccess}</span>
		</div>
	{/if}
	{#if formError}
		<div class="message error-message">
			<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="20" height="20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" /></svg>
			<span>{formError}</span>
		</div>
	{/if}

	<form
		class="recipe-form"
		method="POST"
		enctype="multipart/form-data"
		on:submit|preventDefault={async (event) => {
			isLoading = true;
			formError = null;
			formSuccess = null;

			const formData = new FormData();
			formData.append('name', recipeName);
			formData.append('ingredients', ingredientsStr);
			formData.append('method', method);
			if (photoFile && photoFile.length > 0) {
				formData.append('photo', photoFile[0]);
			}

			try {
				const response = await fetch('/api/v1/recipes', {
					method: 'POST',
					body: formData,
				});

				isLoading = false;
				if (response.ok) {
					const createdRecipe: Recipe = await response.json();
					formSuccess = 'Recipe created successfully! Taking you there...';
					if (createdRecipe && createdRecipe.id) {
						setTimeout(() => goto(`/recipes/${createdRecipe.id}`), 1500); // Delay for success message
					} else {
						// Clear form as a fallback
						recipeName = '';
						ingredientsStr = '';
						method = '';
						const photoInput = document.getElementById('photo') as HTMLInputElement;
						if (photoInput) photoInput.value = '';
					}
				} else {
					const errorData = await response.json();
					formError = errorData.error || `Failed to create recipe. Status: ${response.status}`;
				}
			} catch (err: any) {
				isLoading = false;
				formError = err.message || 'An unexpected network error occurred.';
				console.error('Submission error:', err);
			}
		}}
	>
		<div class="form-group">
			<label for="name" class="form-label">Recipe Name:</label>
			<input type="text" id="name" class="form-input" bind:value={recipeName} required />
		</div>

		<div class="form-group">
			<label for="ingredients" class="form-label">Ingredients:</label>
			<input type="text" id="ingredients" class="form-input" bind:value={ingredientsStr} placeholder="e.g., 1 cup flour, 2 eggs, 1 tsp sugar" />
			<small class="form-hint">Comma-separated, please!</small>
		</div>

		<div class="form-group">
			<label for="method" class="form-label">Method:</label>
			<textarea id="method" class="form-textarea" bind:value={method} rows="10" required placeholder="Describe the cooking steps..."></textarea>
		</div>

		<div class="form-group">
			<label for="photo" class="form-label">Photo (optional):</label>
			<input type="file" id="photo" class="form-input-file" accept="image/*" on:change={(e) => photoFile = (e.currentTarget as HTMLInputElement).files} />
		</div>

		<div class="form-actions">
			<a href="/" class="button secondary cancel-button">Cancel</a>
			<button type="submit" class="button primary submit-button" disabled={isLoading}>
				{#if isLoading}
					<span class="spinner"></span> Creating...
				{:else}
					<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="18" height="18" style="margin-right: 8px;"><path d="M10.75 4.75a.75.75 0 00-1.5 0v4.5h-4.5a.75.75 0 000 1.5h4.5v4.5a.75.75 0 001.5 0v-4.5h4.5a.75.75 0 000-1.5h-4.5v-4.5z" /></svg>
					Create Recipe
				{/if}
			</button>
		</div>
	</form>
</div>

<style>
	.form-page-container { /* Extends .main-container for form specific layout */
		max-width: 700px; /* Slightly narrower for forms */
	}

	.page-title {
		text-align: center;
		font-size: 2em;
		color: var(--color-primary);
		margin-bottom: 30px;
	}

	.recipe-form {
		background-color: var(--color-surface);
		padding: 25px 30px;
		border-radius: var(--border-radius);
		box-shadow: var(--shadow-md);
	}

	.form-group {
		margin-bottom: 25px;
	}

	.form-label {
		display: block;
		margin-bottom: 8px;
		font-weight: 600;
		font-size: 1.05em;
		color: var(--color-text-light);
	}

	/* Inputs and Textarea use global styles from app.css but can be tweaked */
	.form-input, .form-textarea {
		width: 100%;
		/* Global styles apply for padding, border, border-radius, font-size */
	}
	.form-textarea {
		min-height: 120px;
		resize: vertical;
	}
	.form-input-file {
		/* Basic styling, browser default is often hard to override fully without JS */
		padding: 8px;
		border: 1px solid var(--color-border);
		border-radius: var(--border-radius);
		width: 100%;
		background-color: var(--color-background); /* Light bg for file input */
	}
	.form-input-file:hover {
		border-color: var(--color-primary);
	}


	.form-hint {
		display: block;
		margin-top: 6px;
		font-size: 0.9em;
		color: var(--color-text-light);
	}

	.form-actions {
		margin-top: 30px;
		display: flex;
		justify-content: flex-end; /* Align buttons to the right */
		gap: 15px;
		align-items: center;
	}

	.button { /* General button styling from app.css is base */
		padding: 10px 20px;
		font-size: 1em;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}
	.button.primary { /* From app.css */
		/* background-color: var(--color-primary); color: white; */
	}
	.button.secondary {
		background-color: var(--color-text-light);
		color: white;
	}
	.button.secondary:hover {
		background-color: var(--color-text);
	}
	.submit-button:disabled {
		background-color: #cccccc;
		cursor: not-allowed;
	}
	.submit-button .spinner {
		width: 16px;
		height: 16px;
		border: 2px solid rgba(255,255,255,0.3);
		border-radius: 50%;
		border-top-color: white;
		animation: spin 1s ease-infinite;
		margin-right: 8px;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.message {
		display: flex;
		align-items: center;
		padding: 12px 15px;
		margin-bottom: 20px;
		border-radius: var(--border-radius);
		font-weight: 500;
	}
	.message svg {
		margin-right: 10px;
		flex-shrink: 0; /* Prevent icon from shrinking */
	}
	.success-message {
		background-color: #e6f4ea; /* Light green */
		color: var(--color-primary-dark);
		border: 1px solid var(--color-primary);
	}
	.error-message {
		background-color: #fdecea; /* Light red */
		color: var(--color-error);
		border: 1px solid var(--color-error);
	}

	.cancel-button {
		/* Uses .button.secondary */
	}
</style>