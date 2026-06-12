import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { graphqlRequest } from '@/gql/graphql';
import { useErrorHandler } from '@/hooks/use-error-handler';
import {
  CreateResponseProtectionRuleInput,
  PreviewResponseProtectionRuleInput,
  ResponseProtectionRule,
  ResponseProtectionRuleConnection,
  ResponseProtectionRulePreviewResult,
  UpdateResponseProtectionRuleInput,
  responseProtectionRuleConnectionSchema,
  responseProtectionRulePreviewResultSchema,
  responseProtectionRuleSchema,
} from './schema';

const RULES_QUERY = `
  query GetResponseProtectionRules(
    $first: Int
    $after: Cursor
    $last: Int
    $before: Cursor
    $where: ResponseProtectionRuleWhereInput
    $orderBy: ResponseProtectionRuleOrder
  ) {
    responseProtectionRules(first: $first, after: $after, last: $last, before: $before, where: $where, orderBy: $orderBy) {
      edges {
        node {
          id
          createdAt
          updatedAt
          name
          description
          pattern
          status
          settings {
            action
            replacement
            scopes
          }
        }
        cursor
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      totalCount
    }
  }
`;

const CREATE_RULE_MUTATION = `
  mutation CreateResponseProtectionRule($input: CreateResponseProtectionRuleInput!) {
    createResponseProtectionRule(input: $input) {
      id
      createdAt
      updatedAt
      name
      description
      pattern
      status
      settings {
        action
        replacement
        scopes
      }
    }
  }
`;

const UPDATE_RULE_MUTATION = `
  mutation UpdateResponseProtectionRule($id: ID!, $input: UpdateResponseProtectionRuleInput!) {
    updateResponseProtectionRule(id: $id, input: $input) {
      id
      createdAt
      updatedAt
      name
      description
      pattern
      status
      settings {
        action
        replacement
        scopes
      }
    }
  }
`;

const DELETE_RULE_MUTATION = `
  mutation DeleteResponseProtectionRule($id: ID!) {
    deleteResponseProtectionRule(id: $id)
  }
`;

const UPDATE_RULE_STATUS_MUTATION = `
  mutation UpdateResponseProtectionRuleStatus($id: ID!, $status: ResponseProtectionRuleStatus!) {
    updateResponseProtectionRuleStatus(id: $id, status: $status)
  }
`;

const BULK_DELETE_RULES_MUTATION = `
  mutation BulkDeleteResponseProtectionRules($ids: [ID!]!) {
    bulkDeleteResponseProtectionRules(ids: $ids)
  }
`;

const BULK_ENABLE_RULES_MUTATION = `
  mutation BulkEnableResponseProtectionRules($ids: [ID!]!) {
    bulkEnableResponseProtectionRules(ids: $ids)
  }
`;

const BULK_DISABLE_RULES_MUTATION = `
  mutation BulkDisableResponseProtectionRules($ids: [ID!]!) {
    bulkDisableResponseProtectionRules(ids: $ids)
  }
`;

const PREVIEW_RULE_MUTATION = `
  mutation PreviewResponseProtectionRule($input: ResponseProtectionRulePreviewInput!) {
    previewResponseProtectionRule(input: $input) {
      result
      hasMatch
    }
  }
`;

interface QueryRulesArgs {
  first?: number;
  after?: string;
  last?: number;
  before?: string;
  where?: Record<string, any>;
  orderBy?: {
    field: 'CREATED_AT' | 'UPDATED_AT' | 'NAME';
    direction: 'ASC' | 'DESC';
  };
}

export function useQueryResponseProtectionRules(args: QueryRulesArgs) {
  return useQuery({
    queryKey: ['response-protection-rules', args],
    queryFn: async () => {
      const data = await graphqlRequest<{ responseProtectionRules: ResponseProtectionRuleConnection }>(RULES_QUERY, args);
      return responseProtectionRuleConnectionSchema.parse(data.responseProtectionRules);
    },
  });
}

export function useCreateResponseProtectionRule() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { handleError } = useErrorHandler();

  return useMutation({
    mutationFn: async (input: CreateResponseProtectionRuleInput) => {
      try {
        const data = await graphqlRequest<{ createResponseProtectionRule: ResponseProtectionRule }>(CREATE_RULE_MUTATION, { input });
        return responseProtectionRuleSchema.parse(data.createResponseProtectionRule);
      } catch (error) {
        handleError(error, { context: t('responseProtectionRules.dialogs.create.title') });
        throw error;
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['response-protection-rules'] });
      toast.success(t('responseProtectionRules.messages.createSuccess'));
    },
  });
}

export function useUpdateResponseProtectionRule() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { handleError } = useErrorHandler();

  return useMutation({
    mutationFn: async ({ id, input }: { id: string; input: UpdateResponseProtectionRuleInput }) => {
      try {
        const data = await graphqlRequest<{ updateResponseProtectionRule: ResponseProtectionRule }>(UPDATE_RULE_MUTATION, { id, input });
        return responseProtectionRuleSchema.parse(data.updateResponseProtectionRule);
      } catch (error) {
        handleError(error, { context: t('responseProtectionRules.dialogs.edit.title') });
        throw error;
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['response-protection-rules'] });
      toast.success(t('responseProtectionRules.messages.updateSuccess'));
    },
  });
}

export function useDeleteResponseProtectionRule() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { handleError } = useErrorHandler();

  return useMutation({
    mutationFn: async (id: string) => {
      try {
        await graphqlRequest(DELETE_RULE_MUTATION, { id });
      } catch (error) {
        handleError(error, { context: 'Delete Response Protection Rule' });
        throw error;
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['response-protection-rules'] });
      toast.success(t('responseProtectionRules.messages.deleteSuccess'));
    },
  });
}

export function useUpdateResponseProtectionRuleStatus() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { handleError } = useErrorHandler();

  return useMutation({
    mutationFn: async ({ id, status }: { id: string; status: 'enabled' | 'disabled' }) => {
      try {
        await graphqlRequest(UPDATE_RULE_STATUS_MUTATION, { id, status });
      } catch (error) {
        handleError(error, { context: 'Update Response Protection Rule Status' });
        throw error;
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['response-protection-rules'] });
      toast.success(t('responseProtectionRules.messages.statusUpdateSuccess'));
    },
  });
}

export function useBulkDeleteResponseProtectionRules() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { handleError } = useErrorHandler();

  return useMutation({
    mutationFn: async (ids: string[]) => {
      try {
        await graphqlRequest(BULK_DELETE_RULES_MUTATION, { ids });
      } catch (error) {
        handleError(error, { context: 'Bulk Delete Response Protection Rules' });
        throw error;
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['response-protection-rules'] });
      toast.success(t('responseProtectionRules.messages.bulkDeleteSuccess'));
    },
  });
}

export function useBulkEnableResponseProtectionRules() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { handleError } = useErrorHandler();

  return useMutation({
    mutationFn: async (ids: string[]) => {
      try {
        await graphqlRequest(BULK_ENABLE_RULES_MUTATION, { ids });
      } catch (error) {
        handleError(error, { context: 'Bulk Enable Response Protection Rules' });
        throw error;
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['response-protection-rules'] });
      toast.success(t('responseProtectionRules.messages.bulkEnableSuccess'));
    },
  });
}

export function usePreviewResponseProtectionRule() {
  return useMutation({
    mutationFn: async (input: PreviewResponseProtectionRuleInput) => {
      const data = await graphqlRequest<{ previewResponseProtectionRule: ResponseProtectionRulePreviewResult }>(PREVIEW_RULE_MUTATION, { input });
      return responseProtectionRulePreviewResultSchema.parse(data.previewResponseProtectionRule);
    },
  });
}

export function useBulkDisableResponseProtectionRules() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { handleError } = useErrorHandler();

  return useMutation({
    mutationFn: async (ids: string[]) => {
      try {
        await graphqlRequest(BULK_DISABLE_RULES_MUTATION, { ids });
      } catch (error) {
        handleError(error, { context: 'Bulk Disable Response Protection Rules' });
        throw error;
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['response-protection-rules'] });
      toast.success(t('responseProtectionRules.messages.bulkDisableSuccess'));
    },
  });
}
