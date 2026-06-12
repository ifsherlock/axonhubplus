import React, { createContext, useCallback, useContext, useMemo, useState } from 'react';
import { ResponseProtectionRule } from '../data/schema';

type DialogType = 'create' | 'edit' | 'delete' | 'bulkEnable' | 'bulkDisable' | 'bulkDelete' | null;

interface RulesContextType {
  open: DialogType;
  setOpen: (open: DialogType) => void;
  currentRow: ResponseProtectionRule | null;
  setCurrentRow: (row: ResponseProtectionRule | null) => void;
  selectedRules: ResponseProtectionRule[];
  setSelectedRules: (rules: ResponseProtectionRule[]) => void;
  resetRowSelection: (() => void) | null;
  setResetRowSelection: (fn: (() => void) | null) => void;
}

const RulesContext = createContext<RulesContextType | undefined>(undefined);

export function ResponseProtectionRulesProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useState<DialogType>(null);
  const [currentRow, setCurrentRow] = useState<ResponseProtectionRule | null>(null);
  const [selectedRules, setSelectedRules] = useState<ResponseProtectionRule[]>([]);
  const [resetRowSelection, setResetRowSelection] = useState<(() => void) | null>(null);

  const handleSetOpen = useCallback((nextOpen: DialogType) => {
    setOpen(nextOpen);
    if (nextOpen !== 'edit' && nextOpen !== 'delete') {
      setCurrentRow(null);
    }
  }, []);

  const value = useMemo(
    () => ({
      open,
      setOpen: handleSetOpen,
      currentRow,
      setCurrentRow,
      selectedRules,
      setSelectedRules,
      resetRowSelection,
      setResetRowSelection,
    }),
    [open, handleSetOpen, currentRow, selectedRules, resetRowSelection]
  );

  return <RulesContext.Provider value={value}>{children}</RulesContext.Provider>;
}

export function useResponseProtectionRules() {
  const context = useContext(RulesContext);
  if (!context) {
    throw new Error('useResponseProtectionRules must be used within ResponseProtectionRulesProvider');
  }

  return context;
}

export default ResponseProtectionRulesProvider;
