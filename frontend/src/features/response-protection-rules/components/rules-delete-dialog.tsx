import { useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { useResponseProtectionRules } from '../context/rules-context';
import { useDeleteResponseProtectionRule } from '../data/rules';

export function RulesDeleteDialog() {
  const { t } = useTranslation();
  const { open, setOpen, currentRow, resetRowSelection } = useResponseProtectionRules();
  const deleteMutation = useDeleteResponseProtectionRule();

  const handleConfirm = useCallback(async () => {
    if (!currentRow) return;
    await deleteMutation.mutateAsync(currentRow.id);
    setOpen(null);
    resetRowSelection?.();
  }, [currentRow, deleteMutation, resetRowSelection, setOpen]);

  return (
    <AlertDialog open={open === 'delete'} onOpenChange={(isOpen) => !isOpen && setOpen(null)}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t('responseProtectionRules.dialogs.delete.title')}</AlertDialogTitle>
          <AlertDialogDescription>
            {t('responseProtectionRules.dialogs.delete.description', { name: currentRow?.name })}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{t('common.buttons.cancel')}</AlertDialogCancel>
          <AlertDialogAction onClick={handleConfirm} disabled={deleteMutation.isPending} className='bg-destructive text-destructive-foreground hover:bg-destructive/90'>
            {t('common.buttons.delete')}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
