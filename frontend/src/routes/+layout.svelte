<script lang="ts">
	import '../app.css'; // Import global styles
	import { page } from '$app/stores'; // To get current path for active link styling
</script>

<header class="app-header">
	<div class="logo-container">
		<a href="/" class="logo-link">
			<img src="/chef-hat.svg" alt="GoRecipes Logo" class="logo-icon" /> <!-- Placeholder logo -->
			<span class="logo-text">GoRecipes</span>
		</a>
	</div>
	<nav class="main-nav">
		<a href="/" class:active={$page.url.pathname === '/'}>All Recipes</a>
		<a href="/recipes/new" class:active={$page.url.pathname === '/recipes/new'}>Add Recipe</a>
		<!-- Add more links here as needed -->
	</nav>
</header>

<main>
	<slot /> <!-- Page content will be injected here -->
</main>

<footer class="app-footer">
	<p>&copy; {new Date().getFullYear()} GoRecipes. Happy Cooking!</p>
</footer>

<style>
	.app-header {
		background-color: var(--color-surface);
		padding: 15px 30px;
		display: flex;
		justify-content: space-between;
		align-items: center;
		box-shadow: var(--shadow-sm);
		border-bottom: 1px solid var(--color-border);
	}

	.logo-container {
		display: flex;
		align-items: center;
	}

	.logo-link {
		display: flex;
		align-items: center;
		text-decoration: none;
		color: var(--color-primary); /* Logo text color */
	}
	.logo-link:hover {
		color: var(--color-primary-dark);
	}

	.logo-icon {
		height: 32px; /* Adjust as needed */
		width: 32px; /* Adjust as needed */
		margin-right: 10px;
		/* Simple SVG for placeholder - replace with your actual logo */
	}

	.logo-text {
		font-size: 1.5em;
		font-weight: 700;
	}

	.main-nav a {
		margin-left: 20px;
		font-size: 1.1em;
		font-weight: 500;
		color: var(--color-text-light);
		text-decoration: none;
		padding-bottom: 5px; /* For border-bottom effect */
		border-bottom: 2px solid transparent;
		transition: color 0.2s ease, border-bottom-color 0.2s ease;
	}

	.main-nav a:hover {
		color: var(--color-primary);
	}
	
	.main-nav a.active {
		color: var(--color-primary);
		border-bottom-color: var(--color-primary);
		font-weight: 600;
	}

	main {
		/* The .main-container class from app.css can be applied to specific page wrappers if needed,
		   or styles can be applied directly here or in page components. */
		padding-top: 20px; /* Example padding */
		padding-bottom: 60px; /* Space for footer */
	}

	.app-footer {
		background-color: var(--color-text);
		color: var(--color-background);
		text-align: center;
		padding: 20px;
		margin-top: 40px; /* Ensure some space before footer */
		font-size: 0.9em;
	}
</style>