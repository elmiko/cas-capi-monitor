#!/bin/env python3
import argparse
import json
import sys
import plotly.express as px
import pandas as pd


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

    print(f'{len(data)} records found')

    x = []
    y = []
    h = []
    c = []
    for record in data:
        msg = record.get('msg')
        match msg:
            case 'MachineDeployments with scaling annotations':
                x.append(record.get('ts'))
                y.append(record.get('count'))
                h.append(record.get('names'))
                c.append('MD w/ scaling annotations')
            case 'Nodes with true ready condition':
                x.append(record.get('ts'))
                y.append(record.get('count'))
                h.append(record.get('names'))
                c.append('ready nodes')
            case 'Nodes with false or unknown ready condition':
                x.append(record.get('ts'))
                y.append(record.get('count'))
                h.append(record.get('names'))
                c.append('not ready nodes')
            case 'Pods in pending phase':
                x.append(record.get('ts'))
                y.append(record.get('count'))
                h.append(record.get('names'))
                c.append('pending pods')
            case _:
                continue

    df = pd.DataFrame(dict(x=x, y=y, h=h, c=c))
    fig = px.line(df, x='x', y='y', hover_name='h', color='c', title='cas-capi-monitor aggregate data')
    fig.show()


if __name__ == '__main__':
    main()
