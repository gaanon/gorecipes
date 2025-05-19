<script lang="ts">
	// import { खाद्यपदार्थ } from '$lib/stores/culinary'; // Removed unused import
	// import { onMount } from 'svelte'; // Removed unused import

	let fileInput: HTMLInputElement;
	let selectedFile: File | null = null;
	let isImporting = false;
	let feedbackMessage = '';
	let feedbackType: 'success' | 'error' | 'info' = 'info';

	// Response structure from the backend
	interface ImportResponse {
		total_recipes_in_file?: number;
		successfully_imported_count?: number;
		skipped_duplicate_count?: number;
		skipped_malformed_count?: number;
		error_message?: string;
	}

	function handleFileSelect(event: Event) {
		const input = event.target as HTMLInputElement;
		if (input.files && input.files.length > 0) {
			selectedFile = input.files[0];
			feedbackMessage = `Selected file: ${selectedFile.name}`;
			feedbackType = 'info';
		} else {
			selectedFile = null;
			feedbackMessage = 'No file selected.';
			feedbackType = 'info';
		}
	}

	async function handleImport() {
		if (!selectedFile) {
			feedbackMessage = 'Please select a JSON file to import.';
			feedbackType = 'error';
			return;
		}

		isImporting = true;
		feedbackMessage = 'Uploading and processing file...';
		feedbackType = 'info';

		const formData = new FormData();
		formData.append('recipes_file', selectedFile);

		try {
			const response = await fetch('/api/v1/admin/import', {
				method: 'POST',
				body: formData
				// Headers are automatically set by FormData for multipart
			});

			const result: ImportResponse = await response.json();

			if (!response.ok) {
				throw new Error(result.error_message || `Import failed with status: ${response.status}`);
			}

			let successMsg = `Import complete!`;
			if (result.total_recipes_in_file !== undefined) {
				successMsg += ` Total recipes in file: ${result.total_recipes_in_file}.`;
			}
			if (result.successfully_imported_count !== undefined) {
				successMsg += ` Successfully imported: ${result.successfully_imported_count}.`;
			}
			if (result.skipped_duplicate_count !== undefined && result.skipped_duplicate_count > 0) {
				successMsg += ` Duplicates skipped: ${result.skipped_duplicate_count}.`;
			}
			if (result.skipped_malformed_count !== undefined && result.skipped_malformed_count > 0) {
				successMsg += ` Malformed/Skipped: ${result.skipped_malformed_count}.`;
			}
			feedbackMessage = successMsg;
			feedbackType = 'success';
		} catch (err: any) {
			console.error('Import error:', err);
			feedbackMessage = err.message || 'An unexpected error occurred during import.';
			feedbackType = 'error';
		} finally {
			isImporting = false;
			// Optionally reset file input
			if (fileInput) {
				fileInput.value = ''; // Reset file input
			}
			selectedFile = null; // Clear selected file state
		}
	}

	// Culinary theme icon (example, adjust as needed)
	// $: importIcon = $ खाद्यपदार्थ?.icons?.upload || '📤'; // Default icon - Reference removed
</script>

<svelte:head>
	<title>Import Recipes - Admin - GoRecipes</title>
</svelte:head>

<div class="container">
	<h1>Import Recipes</h1>
	<p>Upload a JSON file containing recipes to import them into the database. The format should match the JSON export from this system.</p>

	<div class="import-form">
		<div class="file-input-container">
			<label for="recipes-file-input" class="file-label">
				<!-- {@html importIcon} -->
				Choose JSON File
			</label>
			<input
				type="file"
				id="recipes-file-input"
				bind:this={fileInput}
				on:change={handleFileSelect}
				accept=".json,application/json"
				disabled={isImporting}
			/>
			{#if selectedFile && !isImporting}
				<span class="selected-file-name">Selected: {selectedFile.name}</span>
			{/if}
		</div>

		<button on:click={handleImport} disabled={isImporting || !selectedFile} class="import-button">
			{#if isImporting}
				<span class="spinner small-spinner"></span> Importing...
			{:else}
				Upload and Import Recipes
			{/if}
		</button>
	</div>

	{#if feedbackMessage}
		<div class="feedback-message {feedbackType}">
			<p>{feedbackMessage}</p>
		</div>
	{/if}

	<div class="admin-actions">
		<a href="/admin" class="button secondary">Back to Admin Area</a>
	</div>
</div>

<style>
	.container {
		max-width: 700px;
		margin: 20px auto;
		padding: 25px;
		background-color: var(--color-surface);
		border-radius: var(--border-radius-md);
		box-shadow: var(--shadow-lg);
	}

	h1 {
		color: var(--color-primary);
		margin-bottom: 15px;
		text-align: center;
		font-size: 1.8em;
	}
	p {
		margin-bottom: 20px;
		text-align: center;
		color: var(--color-text-light);
		font-size: 0.95em;
	}

	.import-form {
		display: flex;
		flex-direction: column;
		gap: 20px;
		margin-bottom: 25px;
		padding: 20px;
		border: 1px dashed var(--color-border-strong);
		border-radius: var(--border-radius-sm);
		background-color: var(--color-background);
	}

	.file-input-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 10px;
	}

	input[type='file'] {
		display: none; /* Hide default input */
	}

	.file-label {
		display: inline-block;
		padding: 10px 20px;
		background-color: var(--color-secondary);
		color: var(--color-text); /* Assuming secondary has good contrast with text color */
		border-radius: var(--border-radius-sm);
		cursor: pointer;
		transition: background-color 0.2s;
		font-weight: 500;
	}
	.file-label:hover {
		background-color: var(--color-secondary-dark);
	}

	.selected-file-name {
		font-style: italic;
		font-size: 0.9em;
		color: var(--color-text-light);
	}

	.import-button {
		padding: 12px 20px;
		font-size: 1em;
		background-color: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--border-radius-sm);
		cursor: pointer;
		transition: background-color 0.2s;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
	}
	.import-button:hover:not(:disabled) {
		background-color: var(--color-primary-dark);
	}
	.import-button:disabled {
		background-color: #ccc;
		cursor: not-allowed;
	}

	.feedback-message {
		padding: 15px;
		margin-top: 20px;
		border-radius: var(--border-radius-sm);
		text-align: center;
		font-weight: 500;
	}
	.feedback-message.success {
		background-color: var(--color-success-light);
		color: var(--color-success-dark);
		border: 1px solid var(--color-success);
	}
	.feedback-message.error {
		background-color: var(--color-error-light);
		color: var(--color-error-dark);
		border: 1px solid var(--color-error);
	}
	.feedback-message.info {
		background-color: var(--color-info-light);
		color: var(--color-info-dark);
		border: 1px solid var(--color-info);
	}

	.admin-actions {
		margin-top: 30px;
		text-align: center;
	}
	.admin-actions .button.secondary {
		background-color: var(--color-text-light);
		color: white;
	}
	.admin-actions .button.secondary:hover {
		background-color: var(--color-text);
	}

	/* Spinner (re-use if global, or define locally) */
	.spinner.small-spinner {
		width: 16px;
		height: 16px;
		border: 2px solid rgba(255,255,255,0.3);
		border-radius: 50%;
		border-top-color: white;
		animation: spin 1s linear infinite;
		display: inline-block; /* Or flex item if needed */
	}
	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>