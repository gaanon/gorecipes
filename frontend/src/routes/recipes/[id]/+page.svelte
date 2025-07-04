<script lang="ts">
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';
	import type { Comment, Recipe } from '$lib/types';
	import { onMount } from 'svelte';

	export let data: PageData;

	let comments: Comment[] = [];
	let commentsLoading = true;
	let commentsError: string | null = null;

	let newCommentAuthor = '';
	let newCommentContent = '';
	let isSubmittingComment = false;
	let commentSubmissionError: string | null = null;

	let editingCommentId: string | null = null;
	let editingCommentContent: string = '';
	let editingCommentAuthor: string = '';
	let isUpdatingComment = false;
	let commentUpdateError: string | null = null;

	let isDeletingComment = false;
	let commentDeleteError: string | null = null;

	$: recipe = data.recipe as Recipe | null;
	$: error = data.error as string | null; // Error from loading the page
	let deleteError: string | null = null; // Specific error for delete operation
	let isDeleting = false;

	const baseImageUrl = '/uploads/images/';
	$: imageUrl = recipe?.photo_filename ? `${baseImageUrl}${recipe.photo_filename}` : '';

	async function handleDelete() {
		if (!recipe || !recipe.id) return;

		const confirmed = window.confirm(`Are you sure you want to delete the recipe "${recipe.name}"? This action cannot be undone.`);
		if (!confirmed) {
			return;
		}

		isDeleting = true;
		deleteError = null;

		try {
			const response = await fetch(`/api/v1/recipes/${recipe.id}`, {
				method: 'DELETE',
			});

			if (response.ok) {
				// Successfully deleted
				await goto('/'); // Navigate to homepage
			} else {
				const errorData = await response.json();
				deleteError = errorData.error || `Failed to delete recipe. Status: ${response.status}`;
				console.error('Delete error:', deleteError);
			}
		} catch (err: any) {
			deleteError = err.message || 'An unexpected network error occurred during deletion.';
			console.error('Network error during delete:', err);
		} finally {
			isDeleting = false;
		}
	}

	async function fetchComments() {
		if (!recipe || !recipe.id) return;

		commentsLoading = true;
		commentsError = null;
		try {
			const response = await fetch(`/api/v1/recipes/${recipe.id}/comments`);
			if (response.ok) {
				comments = await response.json();
			} else {
				const errorData = await response.json();
				commentsError = errorData.error || `Failed to fetch comments. Status: ${response.status}`;
			}
		} catch (err: any) {
			commentsError = err.message || 'An unexpected network error occurred while fetching comments.';
		} finally {
			commentsLoading = false;
		}
	}

	async function handleSubmitComment() {
		if (!recipe || !recipe.id || !newCommentAuthor || !newCommentContent) return;

		isSubmittingComment = true;
		commentSubmissionError = null;

		try {
			const response = await fetch(`/api/v1/recipes/${recipe.id}/comments`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ author: newCommentAuthor, content: newCommentContent })
			});

			if (response.ok) {
				newCommentAuthor = '';
				newCommentContent = '';
				await fetchComments(); // Refresh comments after successful submission
			} else {
				const errorData = await response.json();
				commentSubmissionError = errorData.error || `Failed to submit comment. Status: ${response.status}`;
			}
		} catch (err: any) {
			commentSubmissionError = err.message || 'An unexpected network error occurred during comment submission.';
		} finally {
			isSubmittingComment = false;
		}
	}

	function startEdit(comment: Comment) {
		editingCommentId = comment.id;
		editingCommentContent = comment.content;
		editingCommentAuthor = comment.author;
	}

	async function handleUpdateComment(commentId: string) {
		if (!editingCommentContent || !editingCommentAuthor) return;

		isUpdatingComment = true;
		commentUpdateError = null;

		try {
			const response = await fetch(`/api/v1/comments/${commentId}`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ author: editingCommentAuthor, content: editingCommentContent })
			});

			if (response.ok) {
				editingCommentId = null;
				editingCommentContent = '';
				editingCommentAuthor = '';
				await fetchComments(); // Refresh comments after successful update
			} else {
				const errorData = await response.json();
				commentUpdateError = errorData.error || `Failed to update comment. Status: ${response.status}`;
			}
		} catch (err: any) {
			commentUpdateError = err.message || 'An unexpected network error occurred during comment update.';
		} finally {
			isUpdatingComment = false;
		}
	}

	async function handleDeleteComment(commentId: string) {
		const confirmed = window.confirm('Are you sure you want to delete this comment?');
		if (!confirmed) return;

		isDeletingComment = true;
		commentDeleteError = null;

		try {
			const response = await fetch(`/api/v1/comments/${commentId}`, {
				method: 'DELETE'
			});

			if (response.ok) {
				await fetchComments(); // Refresh comments after successful deletion
			} else {
				const errorData = await response.json();
				commentDeleteError = errorData.error || `Failed to delete comment. Status: ${response.status}`;
			}
		} catch (err: any) {
			commentDeleteError = err.message || 'An unexpected network error occurred during comment deletion.';
		} finally {
			isDeletingComment = false;
		}
	}

	onMount(() => {
		if (recipe) {
			fetchComments();
		}
	});
</script>

<svelte:head>
	<title>{recipe ? recipe.name : 'Recipe Details'} - GoRecipes</title>
</svelte:head>

<div class="main-container recipe-detail-page">
	{#if recipe}
		<article class="recipe-content-card">
			<h1 class="recipe-title">{recipe.name}</h1>

			<!-- Original image display logic -->
			<div class="recipe-image-container">
				{#if imageUrl}
					<img src={imageUrl} alt="Photo of {recipe.name}" class="recipe-photo" />
				{:else}
					<div class="recipe-photo-placeholder">
						<span>No Image Available</span>
					</div>
				{/if}
			</div>

			<div class="recipe-body">
				<section class="recipe-section ingredients">
					<h2 class="section-title">
						<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="22" height="22"><path d="M10.75 4.75a.75.75 0 00-1.5 0v4.5h-4.5a.75.75 0 000 1.5h4.5v4.5a.75.75 0 001.5 0v-4.5h4.5a.75.75 0 000-1.5h-4.5v-4.5z" /><path fill-rule="evenodd" d="M5.28 3.22a.75.75 0 00-1.06 1.06L6.94 7H6a.75.75 0 000 1.5h.94l-2.72 2.72a.75.75 0 101.06 1.06L8 9.06V10a.75.75 0 001.5 0v-.94l2.72 2.72a.75.75 0 101.06-1.06L10.06 8H11a.75.75 0 000-1.5h-.94l2.72-2.72a.75.75 0 00-1.06-1.06L9 6.94V6a.75.75 0 00-1.5 0v.94L4.78 3.22zM5 16.75A2.75 2.75 0 105 11.25a2.75 2.75 0 000 5.5zM15 16.75a2.75 2.75 0 100-5.5 2.75 2.75 0 000 5.5z" clip-rule="evenodd" /></svg>
						Ingredients
					</h2>
					{#if recipe.ingredients && recipe.ingredients.length > 0}
						<ul class="ingredient-list">
							{#each recipe.ingredients as ingredient}
								<li>{ingredient}</li>
							{/each}
						</ul>
					{:else}
						<p class="empty-state-text">No ingredients listed for this recipe.</p>
					{/if}
				</section>

				<section class="recipe-section method">
					<h2 class="section-title">
						<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="22" height="22"><path fill-rule="evenodd" d="M2 4.75A.75.75 0 012.75 4h14.5a.75.75 0 010 1.5H2.75A.75.75 0 012 4.75zm0 10.5a.75.75 0 01.75-.75h7.5a.75.75 0 010 1.5h-7.5a.75.75 0 01-.75-.75zM2 9.75A.75.75 0 012.75 9h14.5a.75.75 0 010 1.5H2.75A.75.75 0 012 9.75z" clip-rule="evenodd" /></svg>
						Method
					</h2>
					{#if recipe.method}
						<div class="method-text">{@html recipe.method.replace(/\n/g, '<br>')}</div>
					{:else}
						<p class="empty-state-text">No cooking method provided.</p>
					{/if}
				</section>
			</div>

			<div class="recipe-footer">
				<div class="timestamps">
					<p><strong>Created:</strong> {new Date(recipe.created_at).toLocaleString()}</p>
					<p><strong>Updated:</strong> {new Date(recipe.updated_at).toLocaleString()}</p>
				</div>
				<div class="actions-bar">
					<a href="/" class="button secondary back-button">
						<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="18" height="18"><path fill-rule="evenodd" d="M17 10a.75.75 0 01-.75.75H5.612l4.158 3.96a.75.75 0 11-1.04 1.08l-5.5-5.25a.75.75 0 010-1.08l5.5-5.25a.75.75 0 111.04 1.08L5.612 9.25H16.25A.75.75 0 0117 10z" clip-rule="evenodd" /></svg>
						All Recipes
					</a>
					<a href="/recipes/{recipe.id}/edit" class="button edit-button">
						<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="18" height="18"><path d="M2.695 14.763l-1.262 3.154a.5.5 0 00.65.65l3.155-1.262a4 4 0 001.343-.885L17.5 5.5a2.121 2.121 0 00-3-3L3.58 13.42a4 4 0 00-.885 1.343z" /></svg>
						Edit
					</a>
					<button on:click={handleDelete} disabled={isDeleting} class="button danger delete-button">
						{#if isDeleting}
							<span class="spinner small-spinner"></span>Deleting...
						{:else}
							<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="18" height="18"><path fill-rule="evenodd" d="M8.75 1A2.75 2.75 0 006 3.75H4.5a.75.75 0 000 1.5h11a.75.75 0 000-1.5H14A2.75 2.75 0 0011.25 1H8.75zM10 4.75A.75.75 0 0110.75 5.5v7.5a.75.75 0 01-1.5 0v-7.5A.75.75 0 0110 4.75zM4.5 6.5A.75.75 0 015.25 6h9.5a.75.75 0 010 1.5h-9.5A.75.75 0 014.5 6.5z" clip-rule="evenodd" /></svg>
							Delete
						{/if}
					</button>
				</div>
			</div>
			{#if deleteError}
				<div class="message error-message delete-error-feedback">{deleteError}</div>
			{/if}

			<section class="comments-section">
				<h2 class="section-title">
					<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="22" height="22"><path fill-rule="evenodd" d="M18 10c0 3.866-3.582 7-8 7a8.841 8.841 0 01-4.083-.98L2 17l1.336-3.117C2.47 12.118 2 11.104 2 10c0-3.866 3.582-7 8-7s8 3.134 8 7zM7 9H5v2h2V9zm8 0h-2v2h2V9zM9 9h2v2H9V9z" clip-rule="evenodd" /></svg>
					Comments
				</h2>

				{#if commentsLoading}
					<div class="loading-state">
						<span class="spinner small-spinner"></span> Loading comments...
					</div>
				{:else if commentsError}
					<div class="message error-message">{commentsError}</div>
				{:else if comments.length === 0}
					<p class="empty-state-text">No comments yet. Be the first to leave one!</p>
				{:else}
					<ul class="comments-list">
						{#each comments as comment (comment.id)}
							<li class="comment-item">
								{#if editingCommentId === comment.id}
									<div class="edit-comment-form">
										<input type="text" bind:value={editingCommentAuthor} placeholder="Your Name" class="input-field" />
										<textarea bind:value={editingCommentContent} placeholder="Your comment..." class="textarea-field"></textarea>
										<div class="edit-actions">
											<button on:click={() => handleUpdateComment(comment.id)} disabled={isUpdatingComment} class="button primary small">
												{#if isUpdatingComment}
													<span class="spinner small-spinner"></span> Updating...
												{:else}
													Save
												{/if}
											</button>
											<button on:click={() => (editingCommentId = null)} class="button secondary small">Cancel</button>
										</div>
										{#if commentUpdateError}
											<div class="message error-message">{commentUpdateError}</div>
										{/if}
									</div>
								{:else}
									<p class="comment-meta">
										<strong>{comment.author}</strong> on {new Date(comment.created_at).toLocaleString()}
									</p>
									<p class="comment-content">{comment.content}</p>
									<div class="comment-actions">
										<button on:click={() => startEdit(comment)} class="button edit-button small">Edit</button>
										<button on:click={() => handleDeleteComment(comment.id)} disabled={isDeletingComment} class="button danger small">
											{#if isDeletingComment}
												<span class="spinner small-spinner"></span> Deleting...
											{:else}
												Delete
											{/if}
										</button>
									</div>
								{/if}
							</li>
						{/each}
					</ul>
				{/if}

				<div class="new-comment-form-container">
					<h3>Leave a Comment</h3>
					<form on:submit|preventDefault={handleSubmitComment} class="comment-form">
						<div class="form-group">
							<label for="author">Your Name:</label>
							<input type="text" id="author" bind:value={newCommentAuthor} required class="input-field" />
						</div>
						<div class="form-group">
							<label for="content">Your Comment:</label>
							<textarea id="content" bind:value={newCommentContent} required class="textarea-field"></textarea>
						</div>
						<button type="submit" disabled={isSubmittingComment} class="button primary">
							{#if isSubmittingComment}
								<span class="spinner small-spinner"></span> Submitting...
							{:else}
								Submit Comment
							{/if}
						</button>
						{#if commentSubmissionError}
							<div class="message error-message">{commentSubmissionError}</div>
						{/if}
					</form>
				</div>
			</section>

			<!-- Comments Section Removed -->
			<!--
			<section class="comments-container-section">
				<RecipeComments recipeId={recipe.id} />
			</section>
			-->
		</article>

	{:else if error}
		<div class="message error-message full-page-message">
			<h2>Oops! Something went wrong.</h2>
			<p>{error}</p>
			<a href="/" class="button primary">Back to All Recipes</a>
		</div>
	{:else}
		<div class="loading-state full-page-message">
			<span class="spinner large-spinner"></span>
			<h2>Loading Recipe...</h2>
			<p>Please wait a moment.</p>
		</div>
	{/if}
</div>

<style global>
	.recipe-detail-page { /* Extends .main-container */
		padding-bottom: 40px; /* Extra space at bottom */
	}

	.recipe-content-card {
		background-color: var(--color-surface);
		border-radius: var(--border-radius);
		box-shadow: var(--shadow-md);
		overflow: hidden; /* Important for image border radius */
		margin-top: 20px;
	}

	/* .comments-container-section removed */

	.recipe-title {
		font-size: 2.4em;
		font-weight: 700;
		color: var(--color-text);
		text-align: center;
		padding: 25px 20px 15px;
		margin: 0;
		border-bottom: 1px solid var(--color-border);
	}

	.recipe-image-container {
		width: 100%;
		max-height: 450px; /* Max height for the image */
		overflow: hidden; /* Ensure image doesn't break layout */
		background-color: var(--color-border); /* BG for placeholder */
	}
	.recipe-photo {
		width: 100%;
		height: 100%; /* Fill container */
		max-height: 450px; /* Consistent with container */
		object-fit: cover; /* Cover the area, might crop */
		display: block; /* Remove extra space below image */
	}
	.recipe-photo-placeholder {
		width: 100%;
		height: 300px; /* Default height for placeholder */
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--color-text-light);
		font-size: 1.2em;
	}
	.recipe-photo-placeholder span {
		padding: 10px 15px;
		background-color: rgba(0,0,0,0.05);
		border-radius: var(--border-radius);
	}

	.recipe-body {
		padding: 20px 30px; /* More padding for content */
	}

	.recipe-section {
		margin-bottom: 35px;
	}
	.section-title {
		font-size: 1.6em;
		font-weight: 600;
		color: var(--color-primary);
		margin-bottom: 15px;
		padding-bottom: 8px;
		border-bottom: 2px solid var(--color-primary-dark);
		display: flex;
		align-items: center;
	}
	.section-title svg {
		margin-right: 10px;
	}

	.ingredient-list {
		list-style-type: none; /* Remove default bullets */
		padding-left: 0;
	}
	.ingredient-list li {
		padding: 8px 0 8px 25px; /* Space for custom bullet */
		font-size: 1.05em;
		line-height: 1.7;
		color: var(--color-text-light);
		position: relative; /* For custom bullet positioning */
		border-bottom: 1px dashed var(--color-border);
	}
	.ingredient-list li:last-child {
		border-bottom: none;
	}
	.ingredient-list li::before {
		content: '🍳'; /* Fun emoji bullet */
		position: absolute;
		left: 0;
		top: 8px; /* Adjust vertical alignment */
		color: var(--color-secondary); /* Use accent color */
	}

	.method-text {
		font-size: 1.05em;
		line-height: 1.8;
		color: var(--color-text-light);
		white-space: pre-wrap;
	}
	/*
	.method-text ::selection { // Style selected text in method
		background-color: var(--color-secondary);
		color: var(--color-text);
	}
	*/

	.empty-state-text {
		font-style: italic;
		color: var(--color-text-light);
	}
	
	.recipe-footer {
		padding: 20px 30px;
		background-color: var(--color-background); /* Slightly different bg for footer section */
		border-top: 1px solid var(--color-border);
	}

	.timestamps {
		font-size: 0.85em;
		color: var(--color-text-light);
		margin-bottom: 20px;
		text-align: right;
	}
	.timestamps p {
		margin: 3px 0;
	}
	.timestamps strong {
		font-weight: 600;
	}

	.actions-bar {
		display: flex;
		gap: 12px;
		align-items: center;
		flex-wrap: wrap;
		justify-content: flex-start; /* Align to start */
	}

	.button { /* General button styling from app.css is base */
		padding: 10px 18px;
		font-size: 0.95em;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}
	.button svg {
		margin-right: 8px;
	}
	.button.secondary { /* For back button */
		background-color: var(--color-text-light);
		color: white;
	}
	.button.secondary:hover {
		background-color: var(--color-text);
	}
	.edit-button {
		background-color: var(--color-secondary);
		color: var(--color-text);
	}
	.edit-button:hover {
		background-color: #ffca28; /* Darker yellow */
	}
	.delete-button {
		background-color: var(--color-error);
		color: white;
	}
	.delete-button:hover {
		background-color: #d32f2f; /* Darker red */
	}
	.delete-button:disabled {
		background-color: #cccccc;
		cursor: not-allowed;
	}
	.delete-button .spinner { /* Reusing spinner style */
		width: 14px;
		height: 14px;
		border: 2px solid rgba(255,255,255,0.3);
		border-radius: 50%;
		border-top-color: white;
		animation: spin 1s ease-infinite;
		margin-right: 8px;
	}
	@keyframes spin { to { transform: rotate(360deg); } }


	.message { /* For general error from page load or delete feedback */
		display: flex;
		align-items: center;
		padding: 12px 15px;
		margin-top: 20px;
		border-radius: var(--border-radius);
		font-weight: 500;
	}
	/* .message svg { margin-right: 10px; flex-shrink: 0; } */
	.error-message {
		background-color: #fdecea;
		color: var(--color-error);
		border: 1px solid var(--color-error);
	}
	.delete-error-feedback { /* Specific class if needed, or combine with .error-message */
		margin-top: 15px; /* Ensure spacing if it's separate */
	}

	.full-page-message { /* For loading/error states that take over the page */
		text-align: center;
		padding: 40px 20px;
		margin-top: 30px;
		background-color: var(--color-surface);
		border-radius: var(--border-radius);
		box-shadow: var(--shadow-sm);
	}
	.full-page-message h2 {
		font-size: 1.8em;
		color: var(--color-primary);
		margin-bottom: 10px;
	}
	.full-page-message p {
		font-size: 1.1em;
		margin-bottom: 20px;
	}
	.loading-state .spinner.large-spinner {
		width: 40px;
		height: 40px;
		border-width: 4px;
		margin: 0 auto 20px auto;
		display: block;
		border-top-color: var(--color-primary);
	}

	/* Global styles for form elements, if not already defined in app.css */
	.input-field, .textarea-field {
		width: 100%;
		padding: 10px 12px;
		margin-bottom: 15px;
		border: 1px solid var(--color-border);
		border-radius: var(--border-radius);
		font-size: 1em;
		box-sizing: border-box; /* Include padding and border in the element's total width and height */
		background-color: var(--color-background);
		color: var(--color-text);
	}

	.textarea-field {
		min-height: 100px;
		resize: vertical;
	}

	.input-field:focus, .textarea-field:focus {
		border-color: var(--color-primary);
		outline: none;
		box-shadow: 0 0 0 2px rgba(var(--color-primary-rgb), 0.2);
	}

	.form-group label {
		display: block;
		margin-bottom: 8px;
		font-weight: 600;
		color: var(--color-text);
	}

	.button.small {
		padding: 6px 12px;
		font-size: 0.85em;
	}

	/* Comments Section Specific Styles */
	.comments-section {
		background-color: var(--color-surface);
		border-radius: var(--border-radius);
		box-shadow: var(--shadow-md);
		margin-top: 20px;
		padding: 25px 30px;
	}

	.comments-list {
		list-style: none;
		padding: 0;
		margin-top: 20px;
	}

	.comment-item {
		border-bottom: 1px solid var(--color-border);
		padding: 15px 0;
	}

	.comment-item:last-child {
		border-bottom: none;
	}

	.comment-meta {
		font-size: 0.9em;
		color: var(--color-text-light);
		margin-bottom: 5px;
	}

	.comment-meta strong {
		color: var(--color-primary);
	}

	.comment-content {
		font-size: 1em;
		line-height: 1.6;
		color: var(--color-text);
		margin-bottom: 10px;
	}

	.comment-actions {
		display: flex;
		gap: 10px;
		margin-top: 10px;
	}

	.new-comment-form-container {
		margin-top: 30px;
		padding-top: 20px;
		border-top: 1px solid var(--color-border);
	}

	.new-comment-form-container h3 {
		font-size: 1.4em;
		color: var(--color-primary);
		margin-bottom: 20px;
	}

	.edit-comment-form {
		background-color: var(--color-background);
		padding: 15px;
		border-radius: var(--border-radius);
		margin-top: 10px;
		margin-bottom: 10px;
	}

	.edit-comment-form .input-field,
	.edit-comment-form .textarea-field {
		margin-bottom: 10px;
	}

	.edit-actions {
		display: flex;
		gap: 10px;
		margin-top: 10px;
	}
</style>