<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { toast } from '@zerodevx/svelte-toast';

  let { onRecipeProcessed, onClose, isOpen }: {
    onRecipeProcessed: (recipe: { name: string; ingredients: string[]; method: string }) => void;
    onClose: () => void;
    isOpen: boolean;
  } = $props();

  let isDragging = false;
  let isProcessing = false;
  let previewUrl: string | null = null;
  let file: File | null = null;
  let error: string | null = null;

  const dispatch = createEventDispatcher();
  let dialog: HTMLDialogElement;

  $effect(() => {
    if (dialog && isOpen) {
      dialog.showModal();
    } else if (dialog && !isOpen) {
      dialog.close();
    }
  });

  function handleDragOver(event: DragEvent) {
    event.preventDefault();
    isDragging = true;
  }

  function handleDragLeave() {
    isDragging = false;
  }

  function handleDrop(event: DragEvent) {
    event.preventDefault();
    isDragging = false;
    
    const files = event.dataTransfer?.files;
    if (files && files.length > 0) {
      handleFile(files[0]);
    }
  }

  function handleFileSelect(event: Event) {
    const target = event.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
      handleFile(target.files[0]);
    }
  }

  function handleFile(selectedFile: File) {
    // Check if the file is an image
    if (!selectedFile.type.startsWith('image/')) {
      error = 'Please upload an image file (JPEG, PNG, etc.)';
      return;
    }

    // Check file size (max 5MB)
    if (selectedFile.size > 5 * 1024 * 1024) {
      error = 'File size should be less than 5MB';
      return;
    }

    file = selectedFile;
    error = null;
    
    // Create preview
    const reader = new FileReader();
    reader.onload = (e) => {
      previewUrl = e.target?.result as string;
    };
    reader.readAsDataURL(selectedFile);
  }

  async function processPhoto() {
    if (!file) return;

    isProcessing = true;
    error = null;

    const formData = new FormData();
    formData.append('photo', file);

    try {
      const response = await fetch('/api/v1/recipes/process-photo', {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to process photo');
      }

      const recipeData = await response.json();
      dispatch('recipeProcessed', recipeData);
      onClose();
      toast.push('Recipe extracted successfully!', { duration: 3000 });
    } catch (err) {
      console.error('Error processing photo:', err);
      error = err instanceof Error ? err.message : 'Failed to process photo';
    } finally {
      isProcessing = false;
    }
  }

  function resetForm() {
    file = null;
    previewUrl = null;
    error = null;
  }
</script>

<dialog class="modal" bind:this={dialog} on:close={onClose}>
	<div class="modal-box">
		<form method="dialog">
			<button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2" on:click={onClose}>✕</button>
		</form>
		<h3 class="font-bold text-lg">Upload Recipe Photo</h3>
		{#if !previewUrl}
			<div
				class="mt-2 border-2 border-dashed border-gray-300 rounded-lg p-6 text-center cursor-pointer hover:border-blue-500 transition-colors {isDragging ? 'border-blue-500 bg-blue-50' : ''}"
				on:dragover={handleDragOver}
				on:dragleave={handleDragLeave}
				on:drop={handleDrop}
				on:click|self={() => document.getElementById('file-upload')?.click()}
			>
				<svg
					class="mx-auto h-12 w-12 text-gray-400"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
					xmlns="http://www.w3.org/2000/svg"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
					></path>
				</svg>
				<p class="mt-1 text-sm text-gray-600">
					<span class="font-medium text-blue-600 hover:text-blue-500">
						Upload a file
					</span>
					or drag and drop
				</p>
				<p class="mt-1 text-xs text-gray-500">
					JPG or PNG up to 5MB
				</p>
				<input
					id="file-upload"
					name="file-upload"
					type="file"
					class="sr-only"
					accept="image/*"
					on:change={handleFileSelect}
				/>
			</div>
		{:else}
			<div class="mt-2">
				<img
					src={previewUrl}
					alt="Recipe preview"
					class="mx-auto max-h-64 object-contain rounded-lg"
				/>
			</div>
		{/if}

		{#if error}
			<p class="mt-2 text-sm text-red-600">{error}</p>
		{/if}

		<div class="modal-action">
			{#if previewUrl}
				<button
					type="button"
					class="btn btn-primary"
					on:click={processPhoto}
					disabled={isProcessing}
				>
					{#if isProcessing}
						<span class="loading loading-spinner"></span>
						Processing...
					{:else}
						Process Recipe
					{/if}
				</button>
				<button
					type="button"
					class="btn"
					on:click={resetForm}
					disabled={isProcessing}
				>
					Change Photo
				</button>
			{/if}
			<button
				type="button"
				class="btn"
				on:click={onClose}
				disabled={isProcessing}
			>
				Cancel
			</button>
		</div>
	</div>
</dialog>

<style>
  .fade-enter-active, .fade-leave-active {
    transition: opacity 150ms;
  }
  .fade-enter, .fade-leave-to {
    opacity: 0;
  }
</style>
