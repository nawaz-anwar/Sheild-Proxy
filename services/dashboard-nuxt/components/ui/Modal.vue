<template>
  <TransitionRoot :show="show" as="template">
    <Dialog as="div" class="relative z-50" @close="emit('close')">
      <TransitionChild
        as="template"
        enter="ease-out duration-300"
        enter-from="opacity-0"
        enter-to="opacity-100"
        leave="ease-in duration-200"
        leave-from="opacity-100"
        leave-to="opacity-0"
      >
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
      </TransitionChild>

      <div class="fixed inset-0 z-10 overflow-y-auto">
        <div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
          <TransitionChild
            as="template"
            enter="ease-out duration-300"
            enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            enter-to="opacity-100 translate-y-0 sm:scale-100"
            leave="ease-in duration-200"
            leave-from="opacity-100 translate-y-0 sm:scale-100"
            leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
          >
            <DialogPanel class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6">
              <div>
                <div v-if="icon" :class="iconBgColor" class="mx-auto flex h-12 w-12 items-center justify-center rounded-full">
                  <component :is="icon" :class="iconColor" class="h-6 w-6" />
                </div>
                <div class="mt-3 text-center sm:mt-5">
                  <DialogTitle as="h3" class="text-lg font-semibold leading-6 text-gray-900">
                    {{ title }}
                  </DialogTitle>
                  <div class="mt-2">
                    <slot />
                  </div>
                </div>
              </div>
              <div class="mt-5 sm:mt-6 flex gap-3">
                <slot name="actions">
                  <UiButton
                    v-if="showCancel"
                    variant="secondary"
                    class="flex-1"
                    @click="emit('close')"
                  >
                    Cancel
                  </UiButton>
                  <UiButton
                    class="flex-1"
                    :loading="loading"
                    @click="emit('confirm')"
                  >
                    {{ confirmText }}
                  </UiButton>
                </slot>
              </div>
            </DialogPanel>
          </TransitionChild>
        </div>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<script setup lang="ts">
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue';

defineProps<{
  show: boolean;
  title: string;
  confirmText?: string;
  showCancel?: boolean;
  loading?: boolean;
  icon?: any;
  iconColor?: string;
  iconBgColor?: string;
}>();

const emit = defineEmits<{
  close: [];
  confirm: [];
}>();
</script>
