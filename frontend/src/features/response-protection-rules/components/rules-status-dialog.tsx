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
import { ResponseProtectionRule } from '../data/schema';
import { useUpdateResponseProtectionRuleStatus } from '../data/rules';

interface RulesStatusDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentRow: ResponseProtectionRule;
}

export function RulesStatusDialog({ open, onOpenChange, currentRow }: RulesStatusDialogProps) {
  const { t } = useTranslation();
  const updateStatusMutation = useUpdateResponseProtectionRuleStatus();
  const newStatus = currentRow.status === 'enabled' ? 'disabled' : 'enabled';

  const handleConfirm = useCallback(async () => {
    await updateStatusMutation.mutateAsync({
      id: currentRow.id,
      status: newStatus,
    });
    onOpenChange(false);
  }, [currentRow.id, newStatus, onOpenChange, updateStatusMutation]);

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t('responseProtectionRules.dialogs.statusChange.title')}</AlertDialogTitle>
          <AlertDialogDescription>
            {t(`responseProtectionRules.dialogs.statusChange.description.${newStatus}`, { name: currentRow.name })}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{t('common.buttons.cancel')}</AlertDialogCancel>
          <AlertDialogAction onClick={handleConfirm} disabled={updateStatusMutation.isPending}>
            {t('common.buttons.confirm')}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
