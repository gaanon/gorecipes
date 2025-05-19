<svelte:head>
	<title>Admin Area - GoRecipes</title>
</svelte:head>

<div class="container">
	<h1>Admin Area</h1>
	<p>Welcome to the admin section. Here you can manage various aspects of the application.</p>

	<nav class="admin-nav">
		<ul>
			<li>
				<a href="/admin/export">Export Data</a>
			</li>
			<li>
				<a href="/admin/import">Import Recipes</a>
			</li>
			<!-- More admin tasks can be added here -->
		</ul>
	</nav>
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
		margin-bottom: 20px;
		text-align: center;
	}

	.admin-nav ul {
		list-style-type: none;
		padding: 0;
	}

	.admin-nav li a {
		display: block;
		padding: 10px 15px;
		margin-bottom: 10px;
		background-color: var(--color-background);
		color: var(--color-primary);
		text-decoration: none;
		border-radius: var(--border-radius-sm);
		border: 1px solid var(--color-border);
		transition: background-color 0.2s ease, color 0.2s ease;
	}

	.admin-nav li a:hover {
		background-color: var(--color-primary); /* Use primary green for background */
		color: var(--color-surface); /* Use white for text (contrasts with green) */
		color: var(--color-surface);
	}
</style>