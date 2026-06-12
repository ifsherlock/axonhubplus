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
import { useBulkDeleteResponseProtectionRules } from '../data/rules';

export function RulesBulkDeleteDialog() {
  const { t } = useTranslation();
  const { open, setOpen, selectedRules, resetRowSelection } = useResponseProtectionRules();
  const mutation = useBulkDeleteResponseProtectionRules();

  const handleConfirm = useCallback(async () => {
    await mutation.mutateAsync(selectedRules.map((rule) => rule.id));
    setOpen(null);
    resetRowSelection?.();
  }, [mutation, resetRowSelection, selectedRules, setOpen]);

  return (
    <AlertDialog open={open === 'bulkDelete'} onOpenChange={(isOpen) => !isOpen && setOpen(null)}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t('responseProtectionRules.dialogs.bulkDelete.title')}</AlertDialogTitle>
          <AlertDialogDescription>
            {t('responseProtectionRules.dialogs.bulkDelete.description', { count: selectedRules.length })}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{t('common.buttons.cancel')}</AlertDialogCancel>
          <AlertDialogAction onClick={handleConfirm} disabled={mutation.isPending} className='bg-destructive text-destructive-foreground hover:bg-destructive/90'>
            {t('common.buttons.delete')}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
