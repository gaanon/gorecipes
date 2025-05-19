<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { addRecipeToMealPlan, getStartOfWeek, formatDateToYYYYMMDD } from '$lib/stores/mealPlannerStore';

	export let recipeId: string;
	export let recipeName: string = 'this recipe'; // Optional: for display
	export let showModal: boolean = false;

	const dispatch = createEventDispatcher();

	let currentModalWeekStartDate = getStartOfWeek(new Date());
	let daysInModalWeek: Date[] = [];
	let isLoading = false;
	let modalFeedback = '';
	let modalFeedbackType: 'success' | 'error' | '' = '';

	$: if (showModal) {
		// Reset to current week when modal is shown
		currentModalWeekStartDate = getStartOfWeek(new Date());
		generateDaysInModalWeek(currentModalWeekStartDate);
		modalFeedback = ''; // Clear previous feedback
		modalFeedbackType = '';
	}

	function generateDaysInModalWeek(startDate: Date) {
		const start = getStartOfWeek(new Date(startDate));
		const newDays: Date[] = [];
		for (let i = 0; i < 7; i++) {
			const day = new Date(start);
			day.setDate(start.getDate() + i);
			newDays.push(day);
		}
		daysInModalWeek = newDays;
	}

	function goToModalPreviousWeek() {
		const newStart = new Date(currentModalWeekStartDate);
		newStart.setDate(currentModalWeekStartDate.getDate() - 7);
		currentModalWeekStartDate = newStart;
		generateDaysInModalWeek(currentModalWeekStartDate);
	}

	function goToModalNextWeek() {
		const newStart = new Date(currentModalWeekStartDate);
		newStart.setDate(currentModalWeekStartDate.getDate() + 7);
		currentModalWeekStartDate = newStart;
		generateDaysInModalWeek(currentModalWeekStartDate);
	}

	async function handleDaySelect(date: Date) {
		if (!recipeId) {
			modalFeedback = 'Error: Recipe ID is missing.';
			modalFeedbackType = 'error';
			return;
		}
		isLoading = true;
		modalFeedback = `Adding ${recipeName} to ${formatDateToYYYYMMDD(date)}...`;
		modalFeedbackType = '';

		try {
			await addRecipeToMealPlan(date, recipeId);
			modalFeedback = `${recipeName} successfully added to ${formatDateToYYYYMMDD(date)}!`;
			modalFeedbackType = 'success';
			setTimeout(() => {
				closeModal(); // Close after a short delay on success
			}, 1500);
		} catch (error: any) {
			modalFeedback = error.message || 'Failed to add recipe to plan.';
			modalFeedbackType = 'error';
		} finally {
			isLoading = false;
		}
	}

	function closeModal() {
		showModal = false;
		dispatch('close');
	}

	function handleKeyDown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			closeModal();
		}
	}
</script>

{#if showModal}
<div
	class="modal-backdrop"
	on:click|self={closeModal}
	role="dialog"
	aria-modal="true"
	aria-labelledby="modal-title"
	tabindex="-1"
	on:keydown={handleKeyDown}
>
	<div class="modal-content">
		<header class="modal-header">
			<h3 id="modal-title">Add "{recipeName}" to Meal Plan</h3>
			<button class="close-button" on:click={closeModal} aria-label="Close modal">&times;</button>
		</header>

		<div class="modal-body">
			<p>Select a date to add this recipe:</p>
			<div class="week-navigation-modal">
				<button on:click={goToModalPreviousWeek} disabled={isLoading}>&lt; Prev</button>
				<span>{daysInModalWeek[0]?.toLocaleDateString(undefined, { month: 'short', year: 'numeric' })} - {daysInModalWeek[6]?.toLocaleDateString(undefined, { month: 'short', day:'numeric', year: 'numeric' })}</span>
				<button on:click={goToModalNextWeek} disabled={isLoading}>Next &gt;</button>
			</div>

			<div class="calendar-view-modal">
				{#each daysInModalWeek as day (formatDateToYYYYMMDD(day))}
					<button
						class="day-cell-modal"
						on:click={() => handleDaySelect(day)}
						disabled={isLoading}
						title={day.toLocaleDateString(undefined, { weekday: 'long', month: 'short', day: 'numeric' })}
					>
						<span class="day-name-modal">{day.toLocaleDateString(undefined, { weekday: 'short' })}</span>
						<span class="day-number-modal">{day.getDate()}</span>
					</button>
				{/each}
			</div>

			{#if modalFeedback}
				<p class="feedback-modal {modalFeedbackType}">{modalFeedback}</p>
			{/if}
		</div>
	</div>
</div>
{/if}

<style>
	.modal-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background-color: rgba(0, 0, 0, 0.6);
		display: flex;
		justify-content: center;
		align-items: center;
		z-index: 2000; /* Higher than planner panel */
	}
	.modal-content {
		background-color: var(--color-surface);
		padding: 25px;
		border-radius: var(--border-radius-lg);
		box-shadow: var(--shadow-xl);
		width: 90%;
		max-width: 450px; /* Adjust as needed */
		position: relative;
	}
	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 20px;
		padding-bottom: 15px;
		border-bottom: 1px solid var(--color-border);
	}
	.modal-header h3 {
		margin: 0;
		font-size: 1.3em;
		color: var(--color-primary);
	}
	.close-button {
		background: none;
		border: none;
		font-size: 1.8em;
		cursor: pointer;
		color: var(--color-text-light);
		padding: 0;
		line-height: 1;
	}

	.modal-body p {
		margin-bottom: 15px;
		font-size: 0.95em;
	}

	.week-navigation-modal {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 15px;
		font-size: 0.9em;
	}
	.week-navigation-modal button {
		padding: 5px 10px;
		background-color: var(--color-secondary-light);
		color: var(--color-secondary-dark);
		border: 1px solid var(--color-secondary);
		border-radius: var(--border-radius-sm);
		cursor: pointer;
	}
	.week-navigation-modal button:hover:not(:disabled) {
		background-color: var(--color-secondary);
		color: var(--color-text);
	}
	.week-navigation-modal button:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}
	.week-navigation-modal span {
		font-weight: 500;
		color: var(--color-text);
	}


	.calendar-view-modal {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		gap: 5px;
		margin-bottom: 20px;
	}
	.day-cell-modal {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 8px 4px;
		border: 1px solid var(--color-border-light, #eee);
		border-radius: var(--border-radius-sm);
		background-color: var(--color-surface);
		cursor: pointer;
		transition: background-color 0.2s;
		font-size: 0.8em;
		min-height: 50px;
	}
	.day-cell-modal:hover:not(:disabled) {
		background-color: var(--color-primary-light);
		border-color: var(--color-primary);
	}
	.day-cell-modal:disabled {
		background-color: var(--color-background-disabled);
		cursor: not-allowed;
		opacity: 0.7;
	}
	.day-name-modal {
		font-size: 0.9em;
		color: var(--color-text-light);
	}
	.day-number-modal {
		font-size: 1.1em;
		font-weight: bold;
		color: var(--color-text);
	}

	.feedback-modal {
		margin-top: 15px;
		padding: 10px;
		border-radius: var(--border-radius-sm);
		text-align: center;
		font-size: 0.9em;
	}
	.feedback-modal.success {
		background-color: var(--color-success-light);
		color: var(--color-success-dark);
	}
	.feedback-modal.error {
		background-color: var(--color-error-light);
		color: var(--color-error-dark);
	}
</style>