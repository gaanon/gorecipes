<script lang="ts">
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';
	import type { Recipe } from '$lib/types';

	export let data: PageData;

	let recipeName = data.recipe?.name || '';
	let ingredientsStr = data.recipe?.ingredients?.join(', ') || '';
	let method = data.recipe?.method || '';
	let currentPhotoFilename = data.recipe?.photo_filename || '';
	let photoFile: FileList | null = null;

	let isLoading = false;
	let formError: string | null = data.pageError || null; // Initialize with potential load error
	let formSuccess: string | null = null;

	const recipeId = data.recipe?.id;
	const baseImageUrl = 'http://localhost:8080/uploads/images/';
	$: currentImageUrl = currentPhotoFilename ? `${baseImageUrl}${currentPhotoFilename}` : '';

	async function handleSubmit() {
		if (!recipeId) {
			formError = 'Recipe ID is missing. Cannot update.';
			return;
		}

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
		// If no new photo is selected, the backend should keep the existing one.
		// If a new photo IS selected, the backend handles replacing it.

		try {
			const response = await fetch(`http://localhost:8080/api/v1/recipes/${recipeId}`, {
				method: 'PUT', // Or 'POST' if your backend expects _method for PUT with FormData
				body: formData,
			});

			isLoading = false;
			if (response.ok) {
				const updatedRecipe: Recipe = await response.json();
				formSuccess = 'Recipe updated successfully! Taking you back...';
				// Update current photo filename in case it changed (e.g., to placeholder or new file)
				currentPhotoFilename = updatedRecipe.photo_filename || '';
				// Add a small delay so user can see the success message
				setTimeout(() => goto(`/recipes/${recipeId}`), 1500);
			} else {
				const errorData = await response.json();
				formError = errorData.error || `Failed to update recipe. Status: ${response.status}`;
			}
		} catch (err: any) {
			isLoading = false;
			formError = err.message || 'An unexpected network error occurred.';
			console.error('Update error:', err);
		}
	}
</script>

<svelte:head>
	<title>Edit: {data.recipe?.name || 'Recipe'} - GoRecipes</title>
</svelte:head>

<div class="main-container form-page-container">
	<h1 class="page-title">Update: {data.recipe?.name || 'Recipe'}</h1>

	{#if formSuccess}
		<div class="message success-message">
			<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="20" height="20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clip-rule="evenodd" /></svg>
			<span>{formSuccess}</span>
		</div>
	{/if}

	{#if formError && !data.recipe} <!-- Show general load error if recipe itself failed to load -->
		<div class="message error-message">
			<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="20" height="20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" /></svg>
			<span>{formError}</span>
		</div>
		<div class="form-actions" style="justify-content: center;">
			<a href="/" class="button secondary">Go to Homepage</a>
			{#if recipeId}
			<a href="/recipes/{recipeId}" class="button" style="background-color: var(--color-secondary); color: var(--color-text);">Back to Recipe</a>
			{/if}
		</div>
	{:else}
		{#if formError} <!-- Show form submission errors -->
			<div class="message error-message">
				<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="20" height="20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" /></svg>
				<span>{formError}</span>
			</div>
		{/if}
		<form class="recipe-form" on:submit|preventDefault={handleSubmit}>
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
				<label class="form-label">Current Photo:</label>
				{#if currentImageUrl}
					<img src={currentImageUrl} alt="Current recipe photo" class="current-photo-preview" />
					<p><small class="form-hint">Filename: {currentPhotoFilename}</small></p>
				{:else}
					<p class="form-hint">No current photo.</p>
				{/if}
				<label for="photo" class="form-label" style="margin-top:15px;">Upload New Photo (optional, replaces current):</label>
				<input type="file" id="photo" class="form-input-file" accept="image/*" on:change={(e) => photoFile = (e.currentTarget as HTMLInputElement).files} />
			</div>
			
			<div class="form-actions">
				<a href="/recipes/{recipeId}" class="button secondary cancel-button">Cancel</a>
				<button type="submit" class="button primary submit-button" disabled={isLoading}>
					{#if isLoading}
						<span class="spinner"></span> Saving...
					{:else}
						<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="18" height="18" style="margin-right: 8px;"><path d="M2.695 14.763l-1.262 3.154a.5.5 0 00.65.65l3.155-1.262a4 4 0 001.343-.885L17.5 5.5a2.121 2.121 0 00-3-3L3.58 13.42a4 4 0 00-.885 1.343z" /></svg>
						Save Changes
					{/if}
				</button>
			</div>
		</form>
	{/if}
</div>

<style>
	/* Styles are very similar to Create New Recipe, so we can reuse/adapt */
	.form-page-container {
		max-width: 700px;
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

	.form-input, .form-textarea {
		width: 100%;
	}
	.form-textarea {
		min-height: 120px;
		resize: vertical;
	}
	.form-input-file {
		padding: 8px;
		border: 1px solid var(--color-border);
		border-radius: var(--border-radius);
		width: 100%;
		background-color: var(--color-background);
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
	
	.current-photo-preview {
		max-width: 150px; /* Smaller preview */
		max-height: 150px;
		border-radius: var(--border-radius);
		margin-bottom: 8px;
		border: 1px solid var(--color-border);
		object-fit: cover;
	}

	.form-actions {
		margin-top: 30px;
		display: flex;
		justify-content: flex-end;
		gap: 15px;
		align-items: center;
	}

	.button {
		padding: 10px 20px;
		font-size: 1em;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}
	.button.primary {
		/* Uses global */
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
		flex-shrink: 0;
	}
	.success-message {
		background-color: #e6f4ea;
		color: var(--color-primary-dark);
		border: 1px solid var(--color-primary);
	}
	.error-message {
		background-color: #fdecea;
		color: var(--color-error);
		border: 1px solid var(--color-error);
	}
</style>