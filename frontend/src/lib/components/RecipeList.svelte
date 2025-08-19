<script lang="ts">
  import { fade } from 'svelte/transition';
  
  export let recipes = [];
  export let isLoading = false;
  export let error = null;
  export let onRecipeClick = (recipe) => {};
</script>

{#if isLoading}
  <div class="flex justify-center py-12">
    <div class="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
  </div>
{:else if error}
  <div class="rounded-md bg-red-50 p-4 mb-6">
    <div class="flex">
      <div class="flex-shrink-0">
        <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
        </svg>
      </div>
      <div class="ml-3">
        <h3 class="text-sm font-medium text-red-800">Error loading recipes</h3>
        <div class="mt-2 text-sm text-red-700">
          <p>{error}</p>
        </div>
      </div>
    </div>
  </div>
{:else if recipes.length === 0}
  <div class="text-center py-12">
    <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
    </svg>
    <h3 class="mt-2 text-sm font-medium text-gray-900">No recipes yet</h3>
    <p class="mt-1 text-sm text-gray-500">Get started by adding your first recipe.</p>
  </div>
{:else}
  <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
    {#each recipes as recipe (recipe.id)}
      <div 
        transition:fade
        class="bg-white overflow-hidden shadow rounded-lg hover:shadow-md transition-shadow duration-200 cursor-pointer"
        on:click={() => onRecipeClick(recipe)}
      >
        <div class="px-4 py-5 sm:p-6">
          <h3 class="text-lg font-medium text-gray-900 truncate">{recipe.name}</h3>
          <div class="mt-2 text-sm text-gray-500">
            {#if recipe.ingredients && recipe.ingredients.length > 0}
              <p class="truncate">{recipe.ingredients.length} ingredients</p>
            {/if}
            {#if recipe.prep_time || recipe.cook_time}
              <p class="mt-1 text-xs text-gray-400">
                {recipe.prep_time ? `Prep: ${recipe.prep_time} min` : ''} 
                {recipe.cook_time ? `• Cook: ${recipe.cook_time} min` : ''}
              </p>
            {/if}
          </div>
        </div>
      </div>
    {/each}
  </div>
{/if}
