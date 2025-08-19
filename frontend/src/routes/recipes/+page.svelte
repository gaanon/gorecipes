<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { toast } from '@zerodevx/svelte-toast';
  import RecipeList from '$lib/components/RecipeList.svelte';
  import RecipePhotoUploader from '$lib/components/RecipePhotoUploader.svelte';
  import { PlusIcon, PhotoIcon } from '@heroicons/vue/24/outline';

  let showPhotoUploader = false;
  let recipes = [];
  let isLoading = true;
  let error = null;

  // Fetch recipes
  async function fetchRecipes() {
    try {
      isLoading = true;
      const response = await fetch('/api/v1/recipes');
      if (!response.ok) throw new Error('Failed to fetch recipes');
      const data = await response.json();
      recipes = data.recipes || [];
    } catch (err) {
      console.error('Error fetching recipes:', err);
      error = 'Failed to load recipes. Please try again later.';
      toast.push('Failed to load recipes', { duration: 3000 });
    } finally {
      isLoading = false;
    }
  }

  // Handle recipe processed from photo
  function handleRecipeProcessed(recipeData) {
    // Navigate to the new recipe page with the extracted data
    const queryParams = new URLSearchParams({
      name: recipeData.name,
      method: recipeData.method,
      ingredients: recipeData.ingredients.join('\n')
    });
    
    goto(`/recipes/new?${queryParams.toString()}`);
  }

  onMount(() => {
    fetchRecipes();
  });
</script>

<div class="container mx-auto px-4 py-8">
  <div class="flex justify-between items-center mb-8">
    <h1 class="text-3xl font-bold text-gray-900">My Recipes</h1>
    <div class="flex space-x-4">
      <button
        on:click={() => (showPhotoUploader = true)}
        class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
      >
        <PhotoIcon class="-ml-1 mr-2 h-5 w-5" />
        Upload Photo
      </button>
      <a
        href="/recipes/new"
        class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500"
      >
        <PlusIcon class="-ml-1 mr-2 h-5 w-5" />
        New Recipe
      </a>
    </div>
  </div>

  {#if error}
    <div class="bg-red-50 border-l-4 border-red-400 p-4 mb-6">
      <div class="flex">
        <div class="flex-shrink-0">
          <svg class="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="ml-3">
          <p class="text-sm text-red-700">{error}</p>
        </div>
      </div>
    </div>
  {/if}

  {#if isLoading}
    <div class="flex justify-center items-center py-12">
      <div class="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
    </div>
  {:else}
    {#if recipes.length > 0}
      <RecipeList {recipes} />
    {:else}
      <div class="text-center py-12">
        <svg
          class="mx-auto h-12 w-12 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z"
          />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900">No recipes yet</h3>
        <p class="mt-1 text-sm text-gray-500">
          Get started by creating a new recipe or uploading a photo of one.
        </p>
        <div class="mt-6">
          <button
            on:click={() => (showPhotoUploader = true)}
            type="button"
            class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            <PhotoIcon class="-ml-1 mr-2 h-5 w-5" />
            Upload Photo
          </button>
        </div>
      </div>
    {/if}
  {/if}
</div>

<RecipePhotoUploader 
  isOpen={showPhotoUploader}
  onClose={() => (showPhotoUploader = false)}
  onRecipeProcessed={handleRecipeProcessed}
/>

<style>
  .container {
    max-width: 1200px;
    margin: 0 auto;
  }
</style>
