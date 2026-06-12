import { createFileRoute } from '@tanstack/react-router';
import ResponseProtectionRulesManagement from '@/features/response-protection-rules';

export const Route = createFileRoute('/_authenticated/response-protection-rules/')({
  component: ResponseProtectionRulesManagement,
});
