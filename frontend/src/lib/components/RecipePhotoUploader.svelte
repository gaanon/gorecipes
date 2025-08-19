<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { toast } from '@zerodevx/svelte-toast';

  export let isOpen = false;

  const dispatch = createEventDispatcher();

  let isDragging = false;
  let isProcessing = false;
  let previewUrl: string | null = null;
  let file: File | null = null;
  let error: string | null = null;
  let dialog: HTMLDialogElement;

  onMount(() => {
    if (isOpen) {
      dialog.showModal();
    }
  });

  $: if (dialog) {
    if (isOpen) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }

  function handleClose() {
    isOpen = false;
    resetForm();
    dispatch('close');
  }

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
    if (!selectedFile.type.startsWith('image/')) {
      error = 'Please upload an image file (JPEG, PNG, etc.)';
      return;
    }
    if (selectedFile.size > 10 * 1024 * 1024) {
      error = 'File size should be less than 10MB';
      return;
    }
    file = selectedFile;
    error = null;
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
      handleClose();
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
    const fileInput = document.getElementById('file-upload') as HTMLInputElement;
    if(fileInput) fileInput.value = '';
  }
</script>

<dialog class="modal" bind:this={dialog} on:close={handleClose}>
  <div class="modal-box">
    <form method="dialog">
      <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2" on:click={handleClose}>✕</button>
    </form>
    <h3 class="font-bold text-lg">Upload Recipe Photo</h3>
    
    {#if !previewUrl}
      <label
        for="file-upload"
        class="mt-2 border-2 border-dashed border-gray-300 rounded-lg p-6 text-center cursor-pointer hover:border-blue-500 transition-colors {isDragging ? 'border-blue-500 bg-blue-50' : ''} block"
        on:dragover={handleDragOver}
        on:dragleave={handleDragLeave}
        on:drop={handleDrop}
      >
        <p class="mt-1 text-sm text-gray-600">
          <span class="font-medium text-blue-600 hover:text-blue-500">
            Upload a file
          </span>
          or drag and drop
        </p>
        <p class="mt-1 text-xs text-gray-500">
          JPG or PNG up to 10MB
        </p>
      </label>
      <input
        id="file-upload"
        name="file-upload"
        type="file"
        class="sr-only"
        accept="image/*"
        on:change={handleFileSelect}
      />
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
        on:click={handleClose}
        disabled={isProcessing}
      >
        Cancel
      </button>
    </div>
  </div>
</dialog>
