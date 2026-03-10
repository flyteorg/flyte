import React, { PropsWithChildren } from 'react'
import { Dialog, DialogBackdrop, DialogPanel } from '@headlessui/react'

type ReportDialogProps = {
  isOpen: boolean
  closeDialog: VoidFunction
}

export function ReportDialog({
  isOpen,
  closeDialog,
  children,
}: PropsWithChildren<ReportDialogProps>) {
  return (
    <Dialog open={isOpen} onClose={closeDialog} className="relative z-10">
      <DialogBackdrop className="fixed inset-0 bg-(--system-black)" />
      <div className="fixed inset-0 flex h-screen w-screen items-center justify-center p-3">
        <DialogPanel className="h-full w-full rounded-2xl bg-(--system-white)">
          {children}
        </DialogPanel>
      </div>
    </Dialog>
  )
}
