<svelte:head>
	<title>Export Data - GoRecipes Admin</title>
</svelte:head>

<script lang="ts">
	let isLoading = false;
	let errorMessage: string | null = null;
	let successMessage: string | null = null;
	let exportRecipesOption = true; // Default to exporting recipes
	let exportImagesOption = false;

	async function handleExport() {
		if (!exportRecipesOption && !exportImagesOption) {
			errorMessage = "Please select at least one item to export (recipes or images).";
			successMessage = null;
			return;
		}

		isLoading = true;
		errorMessage = null;
		successMessage = null;

		let requestBody = {
			export_recipes: exportRecipesOption,
			export_images: exportImagesOption
		};
		let defaultFilename = 'export.dat';

		if (exportRecipesOption && exportImagesOption) {
			defaultFilename = 'gorecipes_export.zip';
		} else if (exportRecipesOption) {
			defaultFilename = 'recipes.json';
		} else if (exportImagesOption) {
			defaultFilename = 'recipe_images.zip';
		}


		try {
			const response = await fetch('/api/v1/admin/export', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(requestBody)
			});

			if (!response.ok) {
				const errorData = await response.json().catch(() => ({ error: 'Failed to export data. Server returned an error.' }));
				throw new Error(errorData.error || `HTTP error ${response.status}`);
			}

			const blob = await response.blob();
			const filenameHeader = response.headers.get('Content-Disposition');
			let filename = defaultFilename; // Use dynamic default

			if (filenameHeader) {
				const parts = filenameHeader.split('filename=');
				if (parts.length > 1) {
					filename = parts[1].replace(/"/g, ''); // Remove quotes
				}
			}
			
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = filename;
			document.body.appendChild(a);
			a.click();
			a.remove();
			window.URL.revokeObjectURL(url);

			successMessage = `Successfully exported ${filename}.`;

		} catch (err: any) {
			console.error('Export error:', err);
			errorMessage = err.message || 'An unexpected error occurred during export.';
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="container">
	<h1>Export Data</h1>
	<p>Use this section to export recipe data from the application.</p>

	{#if isLoading}
		<p class="loading-message">Exporting data, please wait...</p>
	{/if}

	{#if successMessage}
		<p class="success-message">{successMessage}</p>
	{/if}

	{#if errorMessage}
		<p class="error-message">Error: {errorMessage}</p>
	{/if}

	<div class="export-options">
		<h2>Export Options</h2>
		
		<div class="option-group">
			<label class="checkbox-label">
				<input type="checkbox" bind:checked={exportRecipesOption} />
				Export Recipes (as .json)
			</label>
			<p class="description">
				Downloads a JSON file containing all recipes.
			</p>
		</div>

		<div class="option-group">
			<label class="checkbox-label">
				<input type="checkbox" bind:checked={exportImagesOption} />
				Export Recipe Images (as .zip)
			</label>
			<p class="description">
				Downloads a ZIP file containing all recipe images. If "Export Recipes" is also selected, images will be included with recipes in a single ZIP.
			</p>
		</div>

		<button
			on:click={handleExport}
			disabled={isLoading || (!exportRecipesOption && !exportImagesOption)}
			class="button primary"
		>
			Start Export
		</button>
		{#if !exportRecipesOption && !exportImagesOption}
			<p class="error-message small-text">Please select at least one option to export.</p>
		{/if}
		
	</div>
	
</div>

<style>
	.container {
		max-width: 800px;
		margin: 20px auto;
		padding: 20px;
		background-color: var(--color-surface);
		border-radius: var(--border-radius-md);
		box-shadow: var(--shadow-md);
	}

	h1 {
		color: var(--color-primary);
		margin-bottom: 10px;
	}
	p {
		margin-bottom: 20px;
		line-height: 1.6;
	}

	.export-options {
		margin-top: 30px;
		padding: 20px;
		border: 1px solid var(--color-border-light);
		border-radius: var(--border-radius-sm);
		background-color: var(--color-background);
	}
	
	.option-group {
		margin-bottom: 20px;
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		cursor: pointer;
		font-size: 1em;
		color: var(--color-text);
	}

	.checkbox-label input[type='checkbox'] {
		margin-right: 10px;
		width: 18px; /* Custom size */
		height: 18px; /* Custom size */
		accent-color: var(--color-primary); /* Color of the checkmark and box when checked */
	}


	.export-options h2 {
		margin-top: 0;
		margin-bottom: 15px;
		color: var(--color-text);
	}
	
	.description {
		font-size: 0.9em;
		color: var(--color-text-light);
		margin-top: 10px;
	}

	.button.primary {
		background-color: var(--color-primary);
		color: var(--color-surface);
		padding: 10px 20px;
		border: none;
		border-radius: var(--border-radius-sm);
		font-size: 1em;
		cursor: pointer;
		transition: background-color 0.2s ease;
	}
	.button.primary:hover:not(:disabled) {
		background-color: var(--color-primary-dark);
	}
	.button.primary:disabled {
		background-color: var(--color-gray-medium);
		cursor: not-allowed;
	}

	.loading-message, .error-message, .success-message {
		padding: 10px;
		margin-bottom: 15px;
		border-radius: var(--border-radius-sm);
		text-align: center;
	}
	.loading-message {
		background-color: var(--color-info-light);
		color: var(--color-info-dark);
		border: 1px solid var(--color-info);
	}
	.success-message {
		background-color: var(--color-success-light);
		color: var(--color-success-dark);
		border: 1px solid var(--color-success);
	}
	.error-message {
		background-color: var(--color-danger-light);
		color: var(--color-danger-dark);
		border: 1px solid var(--color-danger);
	}
	.error-message.small-text {
		font-size: 0.85em;
		padding: 5px;
		margin-top: 5px;
		text-align: left;
	}
</style>