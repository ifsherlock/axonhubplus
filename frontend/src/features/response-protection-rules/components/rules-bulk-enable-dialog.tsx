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
import { useBulkEnableResponseProtectionRules } from '../data/rules';

export function RulesBulkEnableDialog() {
  const { t } = useTranslation();
  const { open, setOpen, selectedRules, resetRowSelection } = useResponseProtectionRules();
  const mutation = useBulkEnableResponseProtectionRules();

  const handleConfirm = useCallback(async () => {
    await mutation.mutateAsync(selectedRules.map((rule) => rule.id));
    setOpen(null);
    resetRowSelection?.();
  }, [mutation, resetRowSelection, selectedRules, setOpen]);

  return (
    <AlertDialog open={open === 'bulkEnable'} onOpenChange={(isOpen) => !isOpen && setOpen(null)}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t('responseProtectionRules.dialogs.bulkEnable.title')}</AlertDialogTitle>
          <AlertDialogDescription>
            {t('responseProtectionRules.dialogs.bulkEnable.description', { count: selectedRules.length })}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{t('common.buttons.cancel')}</AlertDialogCancel>
          <AlertDialogAction onClick={handleConfirm} disabled={mutation.isPending}>
            {t('common.buttons.confirm')}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
