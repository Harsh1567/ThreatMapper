import { useSuspenseQuery } from '@suspensive/react-query';
import { Suspense, useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';
import { createColumnHelper, Table, TableSkeleton } from 'ui-components';

import { DFLink } from '@/components/DFLink';
import { TruncatedText } from '@/components/TruncatedText';
import { SEVERITY_COLORS } from '@/constants/charts';
import { ConnectorHeader } from '@/features/onboard/components/ConnectorHeader';
import { queries } from '@/queries';
import { ScanTypeEnum } from '@/types/common';

const DEFAULT_PAGE_SIZE = 10;

const useGetScanSummary = () => {
  const params = useParams();
  return useSuspenseQuery({
    ...queries.onboard.secretScanSummary({
      scanType: ScanTypeEnum.SecretScan,
      bulkScanId: params.bulkScanId ?? '',
    }),
  });
};

const SummaryTable = () => {
  const { data } = useGetScanSummary();
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const columnHelper = createColumnHelper<(typeof data)[number]>();

  const columns = useMemo(() => {
    const columns = [
      columnHelper.accessor('accountType', {
        cell: (info) => <TruncatedText text={info.getValue()} />,
        header: () => 'Type',
        minSize: 50,
        size: 60,
        maxSize: 60,
      }),
      columnHelper.accessor('accountName', {
        cell: (info) => <TruncatedText text={info.getValue()} />,
        header: () => 'Name',
        minSize: 100,
        size: 120,
        maxSize: 250,
      }),
      columnHelper.accessor('total', {
        cell: (info) => (
          <div className="flex items-center justify-end tabular-nums">
            <span className="truncate">{info.getValue()}</span>
          </div>
        ),
        header: () => (
          <div className="text-right">
            <TruncatedText text="Total" />
          </div>
        ),
        minSize: 80,
        size: 80,
        maxSize: 80,
      }),
      columnHelper.accessor('critical', {
        cell: (info) => {
          return (
            <div className="flex items-center gap-x-2 tabular-nums">
              <div
                className="w-3 h-3 rounded-full"
                style={{
                  backgroundColor: SEVERITY_COLORS['critical'],
                }}
              ></div>
              <span>{info.getValue() ?? 0}</span>
            </div>
          );
        },
        header: () => <TruncatedText text="Critical" />,
        minSize: 80,
        size: 80,
        maxSize: 80,
        enableResizing: false,
      }),
      columnHelper.accessor('high', {
        cell: (info) => {
          return (
            <div className="flex items-center gap-x-2 tabular-nums">
              <div
                className="w-3 h-3 rounded-full shrink-0"
                style={{
                  backgroundColor: SEVERITY_COLORS['high'],
                }}
              ></div>
              <span>{info.getValue() ?? 0}</span>
            </div>
          );
        },
        header: () => <TruncatedText text="High" />,
        minSize: 80,
        size: 80,
        maxSize: 80,
        enableResizing: false,
      }),
      columnHelper.accessor('medium', {
        cell: (info) => {
          return (
            <div className="flex items-center gap-x-2 tabular-nums">
              <div
                className="w-3 h-3 rounded-full shrink-0"
                style={{
                  backgroundColor: SEVERITY_COLORS['medium'],
                }}
              ></div>
              <span>{info.getValue() ?? 0}</span>
            </div>
          );
        },
        header: () => <TruncatedText text="Medium" />,
        minSize: 80,
        size: 80,
        maxSize: 80,
        enableResizing: false,
      }),
      columnHelper.accessor('low', {
        cell: (info) => {
          return (
            <div className="flex items-center gap-x-2 tabular-nums">
              <div
                className="w-3 h-3 rounded-full shrink-0"
                style={{
                  backgroundColor: SEVERITY_COLORS['low'],
                }}
              ></div>
              <span>{info.getValue() ?? 0}</span>
            </div>
          );
        },
        header: () => <TruncatedText text="Low" />,
        minSize: 80,
        size: 80,
        maxSize: 80,
        enableResizing: false,
      }),
      columnHelper.accessor('unknown', {
        cell: (info) => {
          return (
            <div className="flex items-center gap-x-2 tabular-nums">
              <div
                className="w-3 h-3 rounded-full shrink-0"
                style={{
                  backgroundColor: SEVERITY_COLORS['unknown'],
                }}
              ></div>
              <span>{info.getValue() ?? 0}</span>
            </div>
          );
        },
        header: () => <TruncatedText text="Unknown" />,
        minSize: 80,
        size: 80,
        maxSize: 80,
        enableResizing: false,
      }),
    ];

    return columns;
  }, []);

  return (
    <Table
      size="default"
      data={data ?? []}
      columns={columns}
      enableColumnResizing
      enableSorting
      enablePageResize
      pageSize={pageSize}
      enablePagination
      onPageResize={(newSize) => {
        setPageSize(newSize);
      }}
    />
  );
};

const SecretScanSummary = () => {
  return (
    <div className="flex flex-col">
      <ConnectorHeader
        title={'Secret Scan Results Summary'}
        description={'Summary of secret scan result'}
      />

      <DFLink to={'/secret'} unstyled>
        <div className="dark:text-accent-accent dark:hover:text-bg-hover-1 text-p4">
          Go to Secret Dashboard to view details scan result
        </div>
      </DFLink>

      <div className="flex flex-col gap-4 mt-4">
        <Suspense
          fallback={
            <div className="w-full">
              <TableSkeleton columns={8} rows={DEFAULT_PAGE_SIZE} />
            </div>
          }
        >
          <SummaryTable />
        </Suspense>
      </div>
    </div>
  );
};

export const module = {
  element: <SecretScanSummary />,
};
