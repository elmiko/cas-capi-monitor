#!/bin/env python3
import argparse
import json
import sys


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('filename')

    args = parser.parse_args()

    logfile = open(args.filename)
    lines = logfile.readlines()
    data = []
    for l, line in enumerate(lines):
        try:
            d = json.loads(line)
            data.append(d)
        except Exception as ex:
            print(f'error reading log line {l}: {ex}')
            sys.exit(1)

    print(f'{len(data)} records processed')


if __name__ == '__main__':
    main()
